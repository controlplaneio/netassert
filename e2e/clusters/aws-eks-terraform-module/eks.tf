provider "aws" {
  region = var.region
}

data "aws_availability_zones" "available" {}



module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"


  cluster_name    = var.cluster_name
  cluster_version = var.cluster_version

  vpc_id                         = module.vpc.vpc_id
  subnet_ids                     = module.vpc.private_subnets
  cluster_endpoint_public_access = true

  eks_managed_node_group_defaults = {
    ami_type = "AL2_x86_64"
  }

  eks_managed_node_groups = {

    one = {
      name = "${var.node_group_name}1"

      instance_types = ["t3.medium"]

      min_size     = 0
      max_size     = 3
      desired_size = var.desired_size
    }

    two = {
      name = "${var.node_group_name}2"

      instance_types = ["m5.large"]

      min_size     = 0
      max_size     = 3
      desired_size = var.desired_size
    }

  }


  # Extend node-to-node security group rules
  node_security_group_additional_rules = {
    ingress_self_all = {
      description = "Node to node all ports/protocols"
      protocol    = "-1"
      from_port   = 0
      to_port     = 0
      type        = "ingress"
      self        = true
    }
  }
}


resource "null_resource" "generate_kubeconfig" {
  depends_on = [module.eks]

  provisioner "local-exec" {
    command = "aws eks update-kubeconfig --region ${var.region} --name ${module.eks.cluster_name} --kubeconfig ${var.kubeconfig_file}"
  }
}
