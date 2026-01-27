# End-to-End(E2E) Tests

The E2E tests uses `terraform` and `terratest` to spin up GKE and EKS clusters. There are altogether four tests:

- AWS EKS 1.34 with AWS VPC CNI
  - Test AWS EKS 1.34 with the default AWS VPC CNI that support Network Policies
- AWS EKS 1.34 with Calico CNI
  - Test AWS EKS 1.34 with calico CNI v3.35.0. As part of the test, the AWS CNI is uninstalled and Calico is installed
- GCP GKE 1.33 with GCP VPC CNI
  - GKE 1.33 with GCP VPC CNI
- GCP GKE 1.33 GCP Dataplane v2
  - GKE 1.33 with data-plane v2 based on Cilium
- Kind k8s 1.35 with Calico CNI
  - Kind with Calico CNI

*Each test is skipped if the corresponding environment variable is not set.*

| Test                                | Environment variable |
|-------------------------------------|----------------------|
| AWS EKS with AWS VPC CNI            | EKS_VPC_E2E_TESTS    |
| AWS EKS with Calico CNI             | EKS_CALICO_E2E_TESTS |
| GCP GKE with GCP VPC CNI            | GKE_VPC_E2E_TESTS    |
| GCP GKE GCP with DataPlane V2       | GKE_DPV2_E2E_TESTS   |
| Kind with Calico CNI                | KIND_E2E_TESTS       |

## Running tests

- Make sure you have installed `kubectl` and `AWS Cli v2`
- Make sure you have also installed `gke-gcloud-auth-plugin` for kubectl by following this [link](https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke)

For AWS EKS tests, make sure you have valid AWS credentials:

```bash
❯ aws sso login

❯ export EKS_VPC_E2E_TESTS=yes
❯ export EKS_CALICO_E2E_TESTS=yes
```

For GCP GKE tests, make sure you export the Project name and set it as the default project:

```bash
❯ export GOOGLE_PROJECT=<your_project>
❯ gcloud config set project <your_project>
❯ gcloud auth application-default login
❯ export GKE_VPC_E2E_TESTS=yes
❯ export GKE_DPV2_E2E_TESTS=yes
```

Run the tests using the following command:

```bash
# from the root of the project
❯ go test -timeout=91m -v ./e2e/... -count=1
```

Tests can be configured by updating values in [end-to-end test helpers](./helpers/)
