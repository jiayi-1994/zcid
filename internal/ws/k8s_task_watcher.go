package ws

import (
	"context"
	"log/slog"
	"time"

	"github.com/xjy/zcid/internal/stepexec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

var taskRunGVR = schema.GroupVersionResource{Group: "tekton.dev", Version: "v1", Resource: "taskruns"}

func (w *RealK8sWatcher) WatchTaskRuns(ctx context.Context, namespace string) {
	for {
		if err := w.doWatchTaskRuns(ctx, namespace); err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Warn("TaskRun watch error, reconnecting...", slog.String("namespace", namespace), slog.Any("error", err))
			time.Sleep(5 * time.Second)
			continue
		}
		if ctx.Err() != nil {
			return
		}
		time.Sleep(time.Second)
	}
}

func (w *RealK8sWatcher) doWatchTaskRuns(ctx context.Context, namespace string) error {
	watcher, err := w.dynClient.Resource(taskRunGVR).Namespace(namespace).Watch(ctx, metav1.ListOptions{LabelSelector: "zcid.io/managed-by=zcid"})
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
			rows, err := stepexec.ExtractTaskRun(obj)
			if err != nil {
				slog.Error("extract TaskRun step executions failed", slog.String("task_run", obj.GetName()), slog.Any("error", err))
				continue
			}
			if len(rows) == 0 || w.recorder == nil {
				continue
			}
			if err := w.recorder.RecordRows(ctx, rows); err != nil {
				slog.Warn("enqueue TaskRun step executions failed", slog.String("task_run", obj.GetName()), slog.Any("error", err))
			}
		}
	}
}
