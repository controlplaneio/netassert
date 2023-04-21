package helpers

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

type GKECluster struct {
	terraformDir   string
	zone           string
	name           string
	version        string
	networkMode    NetworkMode
	kubeConfig     string
	kubeConfigPath string
	opts           *terraform.Options
}

func NewGKECluster(t *testing.T, terraformDir, clusterNameSuffix string, nm NetworkMode) *GKECluster {
	name := "netassert-" + clusterNameSuffix

	c := &GKECluster{
		terraformDir:   terraformDir,
		zone:           "us-central1-b",
		name:           name,
		version:        "REGULAR",
		networkMode:    nm,
		kubeConfig:     name + ".kubeconfig",
		kubeConfigPath: terraformDir + "/" + name + ".kubeconfig",
	}

	c.opts = terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: c.terraformDir,
		Vars: map[string]interface{}{
			"zone":            c.zone,
			"cluster_name":    c.name,
			"cluster_version": c.version,
			"kubeconfig_file": c.kubeConfig,
			"use_dataplanev2": c.networkMode == DataPlaneV2,
		},
	})
	return c
}

func (g *GKECluster) Create(t *testing.T) {
	// terraform init
	terraform.InitAndPlan(t, g.opts)

	// terraform apply
	terraform.Apply(t, g.opts)
}

func (g *GKECluster) Destroy(t *testing.T) {
	if g.opts != nil {
		terraform.Destroy(t, g.opts)
	}
}

func (g *GKECluster) KubeConfigGet() string {
	return g.kubeConfigPath
}
