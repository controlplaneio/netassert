## End-to-End(E2E) Tests

The E2E tests uses `terraform` and `terratest` to spin up GKE and EKS clusters. There are altogether four tests:

- AWS EKS 1.25 with AWS VPC CNI
  - Test AWS EKS 1.25 with the default AWS VPC CNI
- AWS EKS 1.25 with Calico CNI
  - Test AWS EKS 1.25 with calico CNI v3.35.0. As part of the test, the AWS CNI is uninstalled and Calico is installed
- GCP GKE 1.24 with GCP VPC CNI
  - GKE 1.24 with GCP VPC CNI
- GCP GKE 1.24 GCP Cilium 1.11 (Dataplane v2)
  - GKE 1.24 with data-plane v2 based on Cilium


*Each test is skipped if the corresponding environment variable is not set.*

| Test                                        | Environment variable |
|---------------------------------------------|----------------------|
| AWS EKS 1.25 with AWS VPC CNI               | EKS_VPC_E2E_TESTS    |
| AWS EKS 1.25 with Calico CNI                | EKS_CALICO_E2E_TESTS |
| GCP GKE 1.24 with GCP VPC CNI               | GKE_VPC_E2E_TESTS    |
| GCP GKE 1.24 GCP Cilium 1.11 (DataPlane v2) | GKE_DPV2_E2E_TESTS   |

## Running tests

- Make sure you have installed kubectl and AWS Cli v2
- Make sure you have also installed Install gke-gcloud-auth-plugin for use with kubectl by following this [link](https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke)

For AWS EKS tests, make sure you have valid AWS credentials:

```bash
❯ aws sso login
❯ export AWS_PROFILE=<your_profile_name>
❯ export AWS_REGION=eu-west-1

❯ export EKS_VPC_E2E_TESTS=yes
❯ export EKS_CALICO_E2E_TESTS=yes
```

For GCP GKE tests, make sure you export the Project name:

```bash
❯ export GOOGLE_PROJECT=<your_project>
❯ gcloud auth application-default login
❯ export GKE_VPC_E2E_TESTS=yes
❯ export GKE_DPV2_E2E_TESTS=yes
```

Run the tests using the following command:

```bash
#from the root of the project
❯ go test -timeout=91m -v ./e2e/... -count=1
```
