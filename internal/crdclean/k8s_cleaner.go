package crdclean

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var pipelineRunGVR = schema.GroupVersionResource{
	Group:    "tekton.dev",
	Version:  "v1",
	Resource: "pipelineruns",
}

type RealK8sClient struct {
	dynClient  dynamic.Interface
	namespaces []string
}

func NewRealK8sClient(dynClient dynamic.Interface, namespaces []string) *RealK8sClient {
	return &RealK8sClient{dynClient: dynClient, namespaces: namespaces}
}

func (c *RealK8sClient) DeleteExpiredPipelineRuns(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	var totalDeleted int

	for _, ns := range c.namespaces {
		list, err := c.dynClient.Resource(pipelineRunGVR).Namespace(ns).List(ctx, metav1.ListOptions{
			LabelSelector: "zcid.io/managed-by=zcid",
		})
		if err != nil {
			slog.Warn("failed to list PipelineRuns", slog.String("namespace", ns), slog.Any("error", err))
			continue
		}

		for _, item := range list.Items {
			createdAt := item.GetCreationTimestamp().Time
			if createdAt.Before(cutoff) {
				conditions, found, _ := unstructuredConditions(item.Object)
				if !found || isTerminal(conditions) {
					name := item.GetName()
					if err := c.dynClient.Resource(pipelineRunGVR).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
						slog.Warn("failed to delete expired PipelineRun",
							slog.String("name", name),
							slog.String("namespace", ns),
							slog.Any("error", err),
						)
						continue
					}
					totalDeleted++
					slog.Info("deleted expired PipelineRun",
						slog.String("name", name),
						slog.String("namespace", ns),
						slog.Time("createdAt", createdAt),
					)
				}
			}
		}
	}

	if totalDeleted > 0 {
		slog.Info("CRD cleanup completed", slog.Int("deleted", totalDeleted))
	}
	return nil
}

func unstructuredConditions(obj map[string]interface{}) ([]interface{}, bool, error) {
	status, ok := obj["status"]
	if !ok {
		return nil, false, nil
	}
	statusMap, ok := status.(map[string]interface{})
	if !ok {
		return nil, false, nil
	}
	conditions, ok := statusMap["conditions"]
	if !ok {
		return nil, false, nil
	}
	condSlice, ok := conditions.([]interface{})
	return condSlice, ok, nil
}

func isTerminal(conditions []interface{}) bool {
	if len(conditions) == 0 {
		return false
	}
	last, ok := conditions[len(conditions)-1].(map[string]interface{})
	if !ok {
		return false
	}
	status, _ := last["status"].(string)
	return status == "True" || status == "False"
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.0fd", d.Hours()/24)
}
