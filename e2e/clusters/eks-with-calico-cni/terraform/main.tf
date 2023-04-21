
// we spin up an EKS cluster
module "eks_cluster_calico_cni" {
  source = "../../aws-eks-terraform-module"
  region = var.region
  cluster_name = var.cluster_name
  cluster_version = var.cluster_version
  kubeconfig_file = var.kubeconfig_file
  desired_size = var.desired_size
  node_group_name = var.node_group_name
}


variable "region" {
  description = "AWS region"
  type        = string
}

variable "cluster_version" {
  description = "The AWS EKS cluster version"
  type        = string
}

variable "cluster_name" {
  type        = string
  description = "name of the cluster and VPC"
}

variable "kubeconfig_file" {
  type        = string
  description = "name of the file that contains the kubeconfig information"
  default     = ".kubeconfig"
}


variable "desired_size" {
  type        = number
  description = "desired size of the worker node pool"
  default     = 0
}


variable "node_group_name" {
  type        = string
  description = "prefix of the node group"
  default     = "group"
}
