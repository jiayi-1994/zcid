package ws

import (
	"bufio"
	"context"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type RealLogCollector struct {
	coreClient kubernetes.Interface
}

func NewRealLogCollector(coreClient kubernetes.Interface) *RealLogCollector {
	return &RealLogCollector{coreClient: coreClient}
}

func (c *RealLogCollector) StreamLogs(ctx context.Context, namespace, podName string, handler func(line string)) {
	follow := true
	req := c.coreClient.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow: follow,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		slog.Warn("failed to stream pod logs",
			slog.String("pod", podName),
			slog.String("namespace", namespace),
			slog.Any("error", err),
		)
		return
	}
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			handler(scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Warn("log stream scanner error",
			slog.String("pod", podName),
			slog.Any("error", err),
		)
	}
}
