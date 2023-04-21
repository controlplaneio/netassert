package kubeops

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// generateKubernetsClient - generates a kubernetes client set from default file locations
// first checks KUBECONFIG environment variable
// then the Home directory of the user
// then finally it checks if the program is running in a Pod
func generateKubernetesClient() (kubernetes.Interface, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), nil,
	).ClientConfig()
	if err != nil {
		if restConfig, err := rest.InClusterConfig(); err == nil {
			config = restConfig
		} else {
			return nil, fmt.Errorf("failed to build kubeconfig or InClusterConfig: %w", err)
		}
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return k8sClient, nil
}

// genK8sClientFromKubeConfigFile - Generates kubernetes config file from user supplied kubeConfigPath file
func genK8sClientFromKubeConfigFile(kubeConfigPath string) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig from file %s: %w", kubeConfigPath, err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client from file %s: %w", kubeConfigPath, err)
	}

	return k8sClient, nil
}

// Service exposes the operations on various K8s resources
type Service struct {
	Client kubernetes.Interface // kubernetes client-set
	Log    hclog.Logger         // logger embedded in our service
}

// New - builds a new Service that can interface with Kubernetes
func New(client kubernetes.Interface, l hclog.Logger) *Service {
	return &Service{
		Client: client,
		Log:    l,
	}
}

// NewDefaultService - builds a new Service looking for a KubeConfig in various locations
func NewDefaultService(l hclog.Logger) (*Service, error) {
	clientSet, err := generateKubernetesClient()
	if err != nil {
		return &Service{}, err
	}

	return &Service{
		Client: clientSet,
		Log:    l,
	}, nil
}

// NewServiceFromKubeConfigFile - builds a new Service using KubeConfig file location passed by the caller
func NewServiceFromKubeConfigFile(kubeConfigPath string, l hclog.Logger) (*Service, error) {
	clientSet, err := genK8sClientFromKubeConfigFile(kubeConfigPath)
	if err != nil {
		return &Service{}, err
	}

	return &Service{
		Client: clientSet,
		Log:    l,
	}, nil
}
