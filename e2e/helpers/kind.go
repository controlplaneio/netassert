package helpers

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
)

type KindCluster struct {
	name           string
	networkMode    NetworkMode
	configPath     string
	kubeConfigPath string
	provider       *cluster.Provider
}

func NewKindCluster(t *testing.T, WorkspaceDir string, clusterNameSuffix string, nm NetworkMode) *KindCluster {
	name := "netassert-" + clusterNameSuffix

	c := &KindCluster{
		name:           name,
		networkMode:    nm,
		configPath:     WorkspaceDir + "/kind-config.yaml",
		kubeConfigPath: WorkspaceDir + "/" + name + ".kubeconfig",
	}

	return c
}

func (k *KindCluster) Create(t *testing.T) {
	if _, err := os.Stat(k.configPath); os.IsNotExist(err) {
		t.Fatalf("Error: config file %s does not exit", k.configPath)
	}

	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(cmd.NewLogger()),
	)

	t.Logf("Creating cluster %s", k.name)
	err := provider.Create(
		k.name,
		cluster.CreateWithKubeconfigPath(k.kubeConfigPath),
		cluster.CreateWithConfigFile(k.configPath),
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
	)

	if err != nil {
		t.Fatalf("Error while creating cluster: %v", err)
	}
	k.provider = provider

	options := k8s.NewKubectlOptions("", k.kubeConfigPath, "")
	k8s.KubectlApply(t, options, "https://raw.githubusercontent.com/projectcalico/calico/v3.31.3/manifests/calico.yaml")
}

func (k *KindCluster) Destroy(t *testing.T) {
	if k.provider != nil {
		t.Logf("Deleting cluster %s", k.name)
		if err := k.provider.Delete(k.name, k.kubeConfigPath); err != nil {
			t.Errorf("Error while deleting cluster %s: %v", k.name, err)
		}
	}

	_ = os.Remove(k.kubeConfigPath)
}

func (k *KindCluster) KubeConfigGet() string {
	return k.kubeConfigPath
}

func (k *KindCluster) SkipNetPolTests() bool {
	return false
}
