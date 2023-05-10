package helpers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/controlplaneio/netassert/v2/internal/kubeops"
	"github.com/controlplaneio/netassert/v2/internal/logger"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/terraform"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EKSCluster struct {
	terraformDir   string
	region         string
	name           string
	version        string
	networkMode    NetworkMode
	kubeConfig     string
	kubeConfigPath string
	opts           *terraform.Options
}

func NewEKSCluster(t *testing.T, terraformDir, clusterNameSuffix string, nm NetworkMode) *EKSCluster {
	name := "netassert-" + clusterNameSuffix

	c := &EKSCluster{
		terraformDir:   terraformDir,
		region:         "us-east-2",
		name:           name,
		version:        "1.25",
		networkMode:    nm,
		kubeConfig:     name + ".kubeconfig",
		kubeConfigPath: terraformDir + "/" + name + ".kubeconfig",
	}

	tv := map[string]interface{}{
		"region":          c.region,
		"cluster_version": c.version,
		"cluster_name":    c.name,
		"kubeconfig_file": c.kubeConfig,
		"desired_size":    3,
		"node_group_name": "ng",
	}

	if nm == Calico {
		tv["desired_size"] = 0
		tv["node_group_name"] = "group"
	}

	c.opts = terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: c.terraformDir,
		Vars:         tv,
	})
	return c
}

func (g *EKSCluster) Create(t *testing.T) {
	// terraform init
	terraform.InitAndPlan(t, g.opts)

	// terraform apply
	terraform.Apply(t, g.opts)

	if g.networkMode == Calico {
		g.installCalico(t)
	}
}

func (g *EKSCluster) installCalico(t *testing.T) {
	// once the cluster is ready, we need to follow the instructions here
	// https://docs.tigera.io/calico/3.25/getting-started/kubernetes/managed-public-cloud/eks
	ctx := context.Background()

	lg := logger.NewHCLogger("INFO", "netassertv2-e2e-calico", os.Stdout)

	svc, err := kubeops.NewServiceFromKubeConfigFile(g.kubeConfigPath, lg)
	if err != nil {
		t.Fatalf("Failed to build kubernetes client: %s", err)
	}

	// kubectl delete daemonset -n kube-system aws-node
	err = svc.Client.AppsV1().DaemonSets("kube-system").Delete(
		ctx, "aws-node", metav1.DeleteOptions{})

	if err != nil {
		if !apierrors.IsNotFound(err) {
			t.Fatalf("failed to delete daemonset aws-node in the kube-system namespace")
		}
	}

	svc.Log.Info("AWS-CNI", "msg", "Deleted daemonset aws-node in the kube-system namespace")
	// create a new Kubernetes client using the terratest package
	options := k8s.NewKubectlOptions("", g.kubeConfigPath, "")

	// we now apply calico CNI manifest
	k8s.KubectlApply(t, options, g.terraformDir+"/../calico-3.25.0.yaml")

	// update the desired_size variable to 3
	g.opts.Vars["desired_size"] = 3
	g.opts.Vars["node_group_name"] = "calico"

	newTFOptions := terraform.WithDefaultRetryableErrors(t, g.opts)
	// terraform apply the new options
	// this terraform apply should scale up the worker nodes with Calico CNI
	if _, err := terraform.InitAndApplyE(t, newTFOptions); err != nil {
		t.Fatalf("failed to run terraform init and apply: %s", err)
	}

	svc.Log.Info("Sleeping 20 minutes so connectivity from the cluster to the Internet is restored")
	time.Sleep(20 * time.Minute)
}

func (g *EKSCluster) Destroy(t *testing.T) {
	if g.opts != nil {
		terraform.Destroy(t, g.opts)
	}
}

func (g *EKSCluster) KubeConfigGet() string {
	return g.kubeConfigPath
}

func (g *EKSCluster) SkipNetPolTests() bool {
	if g.networkMode == Calico {
		return false
	}
	return true
}
