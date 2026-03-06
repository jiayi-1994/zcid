package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Clients struct {
	CoreClient    kubernetes.Interface
	DynamicClient dynamic.Interface
	RestConfig    *rest.Config
}

func NewClients() (*Clients, error) {
	cfg, err := getRestConfig()
	if err != nil {
		return nil, fmt.Errorf("build k8s config: %w", err)
	}

	coreClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create core client: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create dynamic client: %w", err)
	}

	return &Clients{
		CoreClient:    coreClient,
		DynamicClient: dynClient,
		RestConfig:    cfg,
	}, nil
}

func getRestConfig() (*rest.Config, error) {
	if _, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST"); ok {
		return rest.InClusterConfig()
	}

	if kubecfg := os.Getenv("KUBECONFIG"); kubecfg != "" {
		return clientcmd.BuildConfigFromFlags("", kubecfg)
	}

	home, _ := os.UserHomeDir()
	if home != "" {
		defPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(defPath); err == nil {
			return clientcmd.BuildConfigFromFlags("", defPath)
		}
	}

	return nil, fmt.Errorf("no kubernetes config found (not in-cluster, no KUBECONFIG env, no ~/.kube/config)")
}

func IsAvailable() bool {
	_, err := getRestConfig()
	return err == nil
}
