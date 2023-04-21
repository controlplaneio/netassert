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
