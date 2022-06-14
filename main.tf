terraform {
  required_version = ">= 0.14.0"
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.35.0"
    }
    selectel = {
      source  = "selectel/selectel"
      version = "~> 3.6.2"
    }
  }
}

provider "openstack" {
  auth_url    = "https://api.selvpc.ru/identity/v3"
  domain_name = var.selectel-account
  tenant_id   = var.project-id
  user_name   = var.openstack-user
  password    = var.openstack-pass
  region      = var.region
}

provider "selectel" {
  token = var.selectel-api-token
}
