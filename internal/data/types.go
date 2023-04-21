package data

import (
	"fmt"
	"io"

	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

// Protocol - represents the Layer 4 protocol
type Protocol string

const (
	// ProtocolTCP - represents the TCP protocol
	ProtocolTCP Protocol = "tcp"

	// ProtocolUDP - represents the UDP protocol
	ProtocolUDP Protocol = "udp"
)

// K8sResourceKind represents the Kind of K8sResource
type K8sResourceKind string

const (
	KindDeployment  K8sResourceKind = "deployment"
	KindStatefulSet K8sResourceKind = "statefulset"
	KindDaemonSet   K8sResourceKind = "daemonset"
	KindPod         K8sResourceKind = "pod"
)

// ValidK8sResourceKinds - holds a map of valid K8sResourceKind
var ValidK8sResourceKinds = map[K8sResourceKind]bool{
	KindDeployment:  true,
	KindStatefulSet: true,
	KindDaemonSet:   true,
	KindPod:         true,
}

// TestType - represents a K8s test type, right now
// we only support k8s type
type TestType string

const (
	K8sTest TestType = "k8s"
)

// TestTypes - holds a map of valid NetAsserTestTypes
var TestTypes = map[TestType]bool{
	K8sTest: true,
}

// K8sResource - Resource hold a Kubernetes Resource
type K8sResource struct {
	Kind      K8sResourceKind `yaml:"kind"`
	Name      string          `yaml:"name"`
	Namespace string          `yaml:"namespace"`
	// Clone     bool            `yaml:"clone"`
}

// Src represents a source in the K8s test
type Src struct {
	K8sResource *K8sResource `yaml:"k8sResource"`
}

// Host represents a host that can be used as Dst in a K8s test
type Host struct {
	Name string `yaml:"name"`
}

// Dst holds the destination or the target resource of the test
type Dst struct {
	K8sResource *K8sResource `yaml:"k8sResource,omitempty"`
	Host        *Host        `yaml:"host,omitempty"`
}

// Test holds a single netAssert test
type Test struct {
	Name           string   `yaml:"name"`
	Type           TestType `yaml:"type"`
	Protocol       Protocol `yaml:"protocol"`
	TargetPort     int      `yaml:"targetPort"`
	TimeoutSeconds int      `yaml:"timeoutSeconds"`
	Attempts       int      `yaml:"attempts"`
	ExitCode       int      `yaml:"exitCode"`
	Src            *Src     `yaml:"src"`
	Dst            *Dst     `yaml:"dst"`
	Pass           bool     `yaml:"pass"`
	FailureReason  string   `yaml:"failureReason"`
}

// Tests - holds a slice of NetAssertTests
type Tests []*Test

func (r *K8sResource) validate() error {
	if r == nil {
		return fmt.Errorf("K8sResource is empty")
	}

	var (
		nameErr         error
		kindErr         error
		nameSpaceErr    error
		resourceKindErr error
	)

	if r.Name == "" {
		nameErr = fmt.Errorf("k8sResource name is missing")
	}

	if r.Kind == "" {
		kindErr = fmt.Errorf("k8sResource kind is missing")
	}

	if r.Namespace == "" {
		nameSpaceErr = fmt.Errorf("k8sResource namespace is missing")
	}

	if _, ok := ValidK8sResourceKinds[r.Kind]; !ok {
		resourceKindErr = fmt.Errorf("k8sResource invalid kind '%s'", r.Kind)
	}

	return multierr.Combine(nameErr, kindErr, nameSpaceErr, resourceKindErr)
}

// validate - validates the Host type
func (h *Host) validate() error {
	if h == nil {
		return fmt.Errorf("host field is nil")
	}

	if h.Name == "" {
		return fmt.Errorf("host field is set to empty string")
	}

	return nil
}

// validate - validates the Dst type
func (d *Dst) validate() error {
	if d == nil {
		return fmt.Errorf("dst field cannot be nil")
	}

	if d.K8sResource != nil && d.Host != nil {
		return fmt.Errorf("dst field only supports K8sResource or Host but not both")
	}

	if d.K8sResource != nil {
		return d.K8sResource.validate()
	}

	if d.Host != nil {
		return d.Host.validate()
	}

	return nil
}

// validate - validates the Src type
func (d *Src) validate() error {
	if d == nil {
		return fmt.Errorf("src field cannot be nil")
	}

	if d.K8sResource == nil {
		return fmt.Errorf("k8sResource field in src is currently the only source allowed")
	}

	return d.K8sResource.validate()
}

// validate - validates the Test case
func (te *Test) validate() error {
	if te == nil {
		return fmt.Errorf("test is pointing to nil")
	}

	var nameErr error
	if te.Name == "" {
		nameErr = fmt.Errorf("name field is missing")
	}

	var invalidProtocolErr error
	if te.Protocol != ProtocolUDP && te.Protocol != ProtocolTCP {
		invalidProtocolErr = fmt.Errorf("invalid protocol %s", te.Protocol)
	}

	var targetPortErr error
	if te.TargetPort < 1 || te.TargetPort > 65535 {
		targetPortErr = fmt.Errorf("targetPort out of range: %d", te.TargetPort)
	}

	var invalidAttemptsErr error
	if te.Attempts < 1 {
		invalidAttemptsErr = fmt.Errorf("attempts must be > 0")
	}

	var timeoutSecondsErr error
	if te.TimeoutSeconds < 1 {
		timeoutSecondsErr = fmt.Errorf("timeoutSeconds must be > 0")
	}

	var invalidTestTypeErr error
	if _, ok := TestTypes[te.Type]; !ok {
		invalidTestTypeErr = fmt.Errorf("invalid test type %v", te.Type)
	}

	var missingSrcErr, k8sResourceErr error
	if te.Src == nil {
		missingSrcErr = fmt.Errorf("src block must be present")
	} else {
		k8sResourceErr = te.Src.validate()
	}

	var missingDstErr, dstValidationErr error
	if te.Dst == nil {
		missingDstErr = fmt.Errorf("dst block must be present")
	} else {
		dstValidationErr = te.Dst.validate()
	}

	var notSupportedTest error
	if te.Protocol == ProtocolUDP && te.Dst != nil && te.Dst.Host != nil {
		notSupportedTest = fmt.Errorf("with udp tests the destination must be a k8sResource")
	}

	return multierr.Combine(nameErr, invalidProtocolErr, targetPortErr,
		invalidAttemptsErr, timeoutSecondsErr, invalidTestTypeErr, k8sResourceErr,
		dstValidationErr, missingSrcErr, missingDstErr, notSupportedTest)
}

// Validate - validates the Tests type
func (ts *Tests) Validate() error {
	testNameMap := make(map[string]struct{})

	for _, test := range *ts {
		if err := test.validate(); err != nil {
			return err
		}

		// if test name already exists
		if _, ok := testNameMap[test.Name]; ok {
			return fmt.Errorf("duplicate test name found %q", test.Name)
		}

		testNameMap[test.Name] = struct{}{}
	}

	return nil
}

// setDefaults - sets sensible defaults to the Test
func (te *Test) setDefaults() {
	if te.TimeoutSeconds == 0 {
		te.TimeoutSeconds = 15
	}

	if te.Attempts == 0 {
		te.Attempts = 3
	}

	if te.Protocol == "" {
		te.Protocol = ProtocolTCP
	}
}

// UnmarshalYAML - decodes Tests type
func (ts *Tests) UnmarshalYAML(node *yaml.Node) error {
	type tmpTests []*Test
	var tmp tmpTests

	if err := node.Decode(&tmp); err != nil {
		return err
	}

	*ts = Tests(tmp)
	if err := ts.Validate(); err != nil {
		return fmt.Errorf("validation failed for tests: %w", err)
	}

	return nil
}

// NewFromReader - creates a new Test from an io.Reader
func NewFromReader(r io.Reader) (Tests, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read from reader: %w", err)
	}

	var tests Tests

	if err := yaml.Unmarshal(buf, &tests); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tests: %w", err)
	}

	if len(tests) == 0 {
		return Tests{}, nil
	}

	return tests, nil
}

// UnmarshalYAML - decodes and validate Test type
func (te *Test) UnmarshalYAML(node *yaml.Node) error {
	// testAlias is an alias to type Test
	// this is need to prevent recursive decoding
	type testAlias Test
	var ta testAlias

	if err := node.Decode(&ta); err != nil {
		return err
	}

	// we need to type cast ta back to p to call the original
	// methods on that type to validate the Test
	p := Test(ta)
	p.setDefaults()
	if err := p.validate(); err != nil {
		return err
	}

	// we need to ensure that te points to the modified type
	*te = p

	return nil
}
