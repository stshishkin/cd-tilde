variable "region" {
  default = "ru-1"
}

variable "az_zone" {
  default = "ru-1b"
}

variable "volume_type" {
  default = "basic.ru-1b"
}

variable "subnet_cidr" {
  default = "10.10.0.0/24"
}

variable "selectel-account" {}

variable "project-id" {}

variable "openstack-user" {}

variable "openstack-pass" {}

variable "selectel-api-token" {}

variable "telegram-bot-token" {}

variable "telegram-chat" {}

variable "countdown-msg" {}

variable "file-msg" {}
