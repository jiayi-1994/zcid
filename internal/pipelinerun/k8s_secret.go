package pipelinerun

import (
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RealSecretInjector struct {
	coreClient kubernetes.Interface
}

func NewRealSecretInjector(coreClient kubernetes.Interface) *RealSecretInjector {
	return &RealSecretInjector{coreClient: coreClient}
}

func (s *RealSecretInjector) InjectSecrets(ctx context.Context, namespace, runID string, secrets map[string]string) (string, error) {
	secretName := "zcid-run-" + runID

	data := make(map[string][]byte, len(secrets))
	for k, v := range secrets {
		data[k] = []byte(v)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"zcid.io/managed-by": "zcid",
				"zcid.io/run-id":     runID,
			},
		},
		Data: data,
	}

	_, err := s.coreClient.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("create secret %s/%s: %w", namespace, secretName, err)
	}

	slog.Info("K8s secret created for pipeline run",
		slog.String("secret", secretName),
		slog.String("namespace", namespace),
		slog.String("runID", runID),
	)
	return secretName, nil
}
