package stepexec

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ExtractTaskRun(obj *unstructured.Unstructured) ([]StepExecution, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil TaskRun")
	}
	runID := strings.TrimSpace(obj.GetLabels()["zcid.io/run-id"])
	if runID == "" {
		return nil, nil
	}
	taskRunName := obj.GetName()
	statusSteps, _, _ := unstructured.NestedSlice(obj.Object, "status", "steps")
	taskSpecSteps, _, _ := unstructured.NestedSlice(obj.Object, "status", "taskSpec", "steps")
	specByName := make(map[string]map[string]interface{}, len(taskSpecSteps))
	for _, raw := range taskSpecSteps {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		if name != "" {
			specByName[name] = m
		}
	}
	results := NewJSONRaw(readNested(obj.Object, []string{"status", "taskResults"}, []any{}))
	if len(statusSteps) == 0 && len(taskSpecSteps) > 0 {
		statusSteps = taskSpecSteps
	}
	rows := make([]StepExecution, 0, len(statusSteps))
	for i, raw := range statusSteps {
		stepStatus, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := stepStatus["name"].(string)
		if name == "" {
			continue
		}
		spec := specByName[name]
		row, err := rowFromStep(obj, taskRunName, runID, name, i, stepStatus, spec, results)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func rowFromStep(obj *unstructured.Unstructured, taskRunName, runID, name string, idx int, st map[string]interface{}, spec map[string]interface{}, results JSONRaw) (StepExecution, error) {
	imageRef := stringPtr(firstString(st, "image", "container"))
	if imageRef != nil && *imageRef == "" {
		imageRef = stringPtr(firstString(spec, "image"))
		if imageRef != nil && *imageRef == "" {
			imageRef = nil
		}
	}
	imageID := firstString(st, "imageID")
	imageDigest := digestFromImageID(imageID)
	started, finished, exitCode, status := stepTiming(st)
	var duration *int64
	if started != nil && finished != nil {
		ms := finished.Sub(*started).Milliseconds()
		if ms < 0 {
			ms = 0
		}
		duration = &ms
	}
	commandArgs := map[string]any{"command": listStrings(spec["command"]), "args": listStrings(spec["args"])}
	envPublic, secretRefs, err := classifyEnv(spec)
	if err != nil {
		return StepExecution{}, err
	}
	row := StepExecution{
		PipelineRunID: runID, TaskRunName: taskRunName, StepName: name, StepIndex: idx, Status: status,
		ImageRef: imageRef, ImageDigest: stringPtr(imageDigest), CommandArgs: NewJSONRaw(commandArgs),
		EnvPublic: NewJSONRaw(envPublic), SecretRefs: NewJSONRaw(secretRefs), ParamsResolved: RawObject(),
		WorkspaceMounts: NewJSONRaw(readMapList(spec, "volumeMounts")), Resources: NewJSONRaw(readNested(spec, []string{"resources"}, map[string]any{})),
		TektonResults: results, OutputDigests: RawArray(), LogRef: NewJSONRaw(map[string]any{"taskRun": taskRunName, "step": name}),
		StartedAt: started, FinishedAt: finished, DurationMS: duration, ExitCode: exitCode,
	}
	if row.ImageDigest != nil && *row.ImageDigest == "" {
		row.ImageDigest = nil
	}
	row.NormalizeJSON()
	ApplySizeLimits(&row)
	return row, nil
}

func stepTiming(st map[string]interface{}) (*time.Time, *time.Time, *int, string) {
	status := StatusPending
	var started, finished *time.Time
	var exit *int
	if running, ok := st["running"].(map[string]interface{}); ok {
		started = parseTimePtr(firstString(running, "startedAt"))
		status = StatusRunning
	}
	if waiting, ok := st["waiting"].(map[string]interface{}); ok && waiting != nil {
		status = StatusPending
	}
	if terminated, ok := st["terminated"].(map[string]interface{}); ok {
		started = parseTimePtr(firstString(terminated, "startedAt"))
		finished = parseTimePtr(firstString(terminated, "finishedAt"))
		if code, ok := numberAsInt(terminated["exitCode"]); ok {
			exit = &code
		}
		reason := strings.ToLower(firstString(terminated, "reason"))
		if exit != nil && *exit == 0 && (reason == "" || reason == "completed") {
			status = StatusSucceeded
		} else if reason == "cancelled" {
			status = StatusCancelled
		} else {
			status = StatusFailed
		}
	}
	return started, finished, exit, status
}

func classifyEnv(spec map[string]interface{}) (map[string]any, []map[string]any, error) {
	envPublic := map[string]any{}
	secretRefs := []map[string]any{}
	if spec == nil {
		return envPublic, secretRefs, nil
	}
	envFrom := readMapList(spec, "envFrom")
	hasEnvFromSecret := false
	for _, src := range envFrom {
		if ref, ok := src["secretRef"].(map[string]interface{}); ok {
			hasEnvFromSecret = true
			secretRefs = append(secretRefs, map[string]any{"name": firstString(ref, "name"), "source": "envFrom.secretRef"})
		}
		if ref, ok := src["configMapRef"].(map[string]interface{}); ok {
			envPublic["envFrom.configMapRef:"+firstString(ref, "name")] = map[string]any{"name": firstString(ref, "name"), "source": "configmap"}
		}
	}
	for _, item := range readMapList(spec, "env") {
		name := firstString(item, "name")
		if vf, ok := item["valueFrom"].(map[string]interface{}); ok {
			for key, source := range map[string]string{"secretKeyRef": "secretKeyRef", "fieldRef": "fieldRef", "resourceFieldRef": "resourceFieldRef"} {
				if ref, ok := vf[key].(map[string]interface{}); ok {
					entry := map[string]any{"env": name, "source": source}
					if refName := firstString(ref, "name"); refName != "" {
						entry["name"] = refName
					}
					if refKey := firstString(ref, "key", "fieldPath", "resource"); refKey != "" {
						entry["key"] = refKey
					}
					secretRefs = append(secretRefs, entry)
				}
			}
			if ref, ok := vf["configMapKeyRef"].(map[string]interface{}); ok {
				envPublic[name] = map[string]any{"name": firstString(ref, "name"), "key": firstString(ref, "key"), "source": "configmap"}
			}
			continue
		}
		if _, ok := item["value"].(string); ok && !hasEnvFromSecret {
			for _, ref := range secretRefs {
				if name != "" && (ref["name"] == name || ref["key"] == name) {
					return nil, nil, fmt.Errorf("literal env %s matches secret ref", name)
				}
			}
			envPublic[name] = map[string]any{"source": "literal", "captured": false}
		}
	}
	return envPublic, secretRefs, nil
}

func firstString(m map[string]interface{}, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := m[k].(string); ok {
			return v
		}
	}
	return ""
}
func stringPtr(s string) *string { return &s }
func digestFromImageID(s string) string {
	if idx := strings.Index(s, "@sha256:"); idx >= 0 {
		return s[idx+1:]
	}
	return ""
}
func parseTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return &t
	}
	return nil
}
func numberAsInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
func listStrings(v any) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		if s, ok := it.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
func readMapList(m map[string]interface{}, key string) []map[string]interface{} {
	arr, ok := m[key].([]interface{})
	if !ok {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(arr))
	for _, it := range arr {
		if mm, ok := it.(map[string]interface{}); ok {
			out = append(out, mm)
		}
	}
	return out
}
func readNested(root map[string]interface{}, path []string, fallback any) any {
	cur := any(root)
	for _, p := range path {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return fallback
		}
		cur, ok = m[p]
		if !ok {
			return fallback
		}
	}
	return cur
}
