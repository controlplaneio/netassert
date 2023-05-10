package helpers

import "testing"

const (
	VPC         NetworkMode = "vpc"
	DataPlaneV2 NetworkMode = "dataplanev2"
	Calico      NetworkMode = "calico"
)

type NetworkMode string

type GenericCluster interface {
	Create(t *testing.T)
	Destroy(t *testing.T)
	KubeConfigGet() string
	SkipNetPolTests() bool
}
