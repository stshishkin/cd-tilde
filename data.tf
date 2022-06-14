
data "openstack_networking_network_v2" "external_net" {
  name = "external-network"
}

data "openstack_images_image_v2" "ubuntu_image" {
  most_recent = true
  visibility  = "public"
  name        = "Ubuntu 20.04 LTS 64-bit"
}
