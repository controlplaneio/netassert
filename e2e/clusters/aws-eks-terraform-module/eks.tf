provider "aws" {
  region = var.region
}

data "aws_availability_zones" "available" {}



module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"
  #version = "~> 19"
  

  cluster_name    = var.cluster_name
  cluster_version = var.cluster_version

  vpc_id                         = module.vpc.vpc_id
  subnet_ids                     = module.vpc.private_subnets
  cluster_endpoint_public_access = true
  enable_cluster_creator_admin_permissions = true

  cluster_addons = {
    vpc-cni = {
      before_compute = true
      most_recent    = true
      configuration_values = jsonencode({
        #resolve_conflicts_on_update = "OVERWRITE"
        enableNetworkPolicy = var.enable_vpc_network_policies ? "true" : "false"
      })
    }
  }

  eks_managed_node_groups = {
    example = {
      name = "${var.node_group_name}1"

      # Starting on 1.30, AL2023 is the default AMI type for EKS managed node groups
      ami_type       = "AL2023_x86_64_STANDARD"
      instance_types = ["t3.medium"]

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
    
    egress_all = {
      description      = "Node all egress"
      protocol         = "-1"
      from_port        = 0
      to_port          = 0
      type             = "egress"
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = ["::/0"]
    }
  }
}

resource "null_resource" "generate_kubeconfig" {
  depends_on = [module.eks]

  provisioner "local-exec" {
    command = "aws eks update-kubeconfig --region ${var.region} --name ${module.eks.cluster_name} --kubeconfig ${var.kubeconfig_file}"
  }
}
