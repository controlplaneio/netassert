
// we spin up an EKS cluster
module "eks_cluster_calico_cni" {
  source = "../../aws-eks-terraform-module"
  region = var.region
  cluster_name = var.cluster_name
  cluster_version = var.cluster_version
  kubeconfig_file = var.kubeconfig_file
  desired_size = var.desired_size
  node_group_name = var.node_group_name
  enable_vpc_network_policies = var.enable_vpc_network_policies
}
