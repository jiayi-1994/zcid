package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

var pipelineRunGVR = schema.GroupVersionResource{
	Group:    "tekton.dev",
	Version:  "v1",
	Resource: "pipelineruns",
}

type RealK8sWatcher struct {
	dynClient dynamic.Interface
}

func NewRealK8sWatcher(dynClient dynamic.Interface) *RealK8sWatcher {
	return &RealK8sWatcher{dynClient: dynClient}
}

func (w *RealK8sWatcher) WatchPipelineRuns(ctx context.Context, namespace string, handler func(runName, status string, stepStatuses []StepStatus)) {
	for {
		if err := w.doWatch(ctx, namespace, handler); err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Warn("PipelineRun watch error, reconnecting...",
				slog.String("namespace", namespace),
				slog.Any("error", err),
			)
			time.Sleep(5 * time.Second)
			continue
		}
		if ctx.Err() != nil {
			return
		}
		time.Sleep(time.Second)
	}
}

func (w *RealK8sWatcher) doWatch(ctx context.Context, namespace string, handler func(runName, status string, stepStatuses []StepStatus)) error {
	watcher, err := w.dynClient.Resource(pipelineRunGVR).Namespace(namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: "zcid.io/managed-by=zcid",
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil
			}
			if event.Type == watch.Error {
				continue
			}
			obj, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}
			name := obj.GetName()
			status, steps := extractStatus(obj)
			handler(name, status, steps)
		}
	}
}

func extractStatus(obj *unstructured.Unstructured) (string, []StepStatus) {
	conditions, found, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	status := "Pending"
	if found && len(conditions) > 0 {
		last, _ := conditions[len(conditions)-1].(map[string]interface{})
		s, _ := last["status"].(string)
		reason, _ := last["reason"].(string)
		switch {
		case s == "True":
			status = "Succeeded"
		case s == "False":
			if reason == "PipelineRunCancelled" {
				status = "Cancelled"
			} else {
				status = "Failed"
			}
		default:
			status = "Running"
		}
	}

	var steps []StepStatus
	taskRuns, found, _ := unstructured.NestedMap(obj.Object, "status", "childReferences")
	if !found {
		childRefs, found, _ := unstructured.NestedSlice(obj.Object, "status", "childReferences")
		if found {
			for _, ref := range childRefs {
				refMap, ok := ref.(map[string]interface{})
				if !ok {
					continue
				}
				pipelineTaskName, _ := refMap["pipelineTaskName"].(string)
				steps = append(steps, StepStatus{
					StepID: pipelineTaskName,
					Status: status,
				})
			}
		}
	} else {
		raw, _ := json.Marshal(taskRuns)
		_ = raw
	}

	return status, steps
}
