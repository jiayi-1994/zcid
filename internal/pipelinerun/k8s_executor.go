package pipelinerun

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/xjy/zcid/pkg/tekton"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var pipelineRunGVR = schema.GroupVersionResource{
	Group:    "tekton.dev",
	Version:  "v1",
	Resource: "pipelineruns",
}

type RealK8sClient struct {
	dynClient dynamic.Interface
}

func NewRealK8sClient(dynClient dynamic.Interface) *RealK8sClient {
	return &RealK8sClient{dynClient: dynClient}
}

func (c *RealK8sClient) SubmitPipelineRun(ctx context.Context, namespace string, pr *tekton.PipelineRun) error {
	data, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("marshal PipelineRun: %w", err)
	}

	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal(data, &obj.Object); err != nil {
		return fmt.Errorf("unmarshal to unstructured: %w", err)
	}

	result, err := c.dynClient.Resource(pipelineRunGVR).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create PipelineRun in K8s: %w", err)
	}
	slog.Info("PipelineRun submitted", slog.String("name", result.GetName()), slog.String("namespace", namespace))
	return nil
}

func (c *RealK8sClient) DeletePipelineRun(ctx context.Context, namespace, name string) error {
	err := c.dynClient.Resource(pipelineRunGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("delete PipelineRun %s/%s: %w", namespace, name, err)
	}
	slog.Info("PipelineRun deleted", slog.String("name", name), slog.String("namespace", namespace))
	return nil
}

func (c *RealK8sClient) GetPipelineRunStatus(ctx context.Context, namespace, name string) (string, error) {
	obj, err := c.dynClient.Resource(pipelineRunGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get PipelineRun %s/%s: %w", namespace, name, err)
	}

	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil || !found || len(conditions) == 0 {
		return "Pending", nil
	}

	last, ok := conditions[len(conditions)-1].(map[string]interface{})
	if !ok {
		return "Unknown", nil
	}

	status, _ := last["status"].(string)
	reason, _ := last["reason"].(string)

	switch {
	case status == "True":
		return "Succeeded", nil
	case status == "False":
		if reason == "PipelineRunCancelled" {
			return "Cancelled", nil
		}
		return "Failed", nil
	default:
		if reason != "" {
			return "Running", nil
		}
		return "Pending", nil
	}
}
