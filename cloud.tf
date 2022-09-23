resource "openstack_compute_keypair_v2" "key_tf" {
  name       = "key_tf-${var.telegram-chat}"
  region     = var.region
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "openstack_networking_router_v2" "router_tf" {
  name                = "router_tf-${var.telegram-chat}"
  external_network_id = data.openstack_networking_network_v2.external_net.id
}

resource "openstack_networking_network_v2" "network_tf" {
  name = "network_tf-${var.telegram-chat}"
}

resource "openstack_networking_subnet_v2" "subnet_tf" {
  network_id = openstack_networking_network_v2.network_tf.id
  name       = "subnet_tf-${var.telegram-chat}"
  cidr       = var.subnet_cidr
}

resource "openstack_networking_router_interface_v2" "router_interface_tf" {
  router_id = openstack_networking_router_v2.router_tf.id
  subnet_id = openstack_networking_subnet_v2.subnet_tf.id
}

resource "random_string" "random_name_server" {
  length  = 16
  special = false
}

resource "openstack_compute_flavor_v2" "flavor_server" {
  name      = "server-${random_string.random_name_server.result}-${var.telegram-chat}"
  ram       = "512"
  vcpus     = "1"
  disk      = "0"
  is_public = "false"
}

resource "openstack_blockstorage_volume_v3" "volume_server" {
  name                 = "volume-for-server1-${var.telegram-chat}"
  size                 = "5"
  image_id             = data.openstack_images_image_v2.ubuntu_image.id
  volume_type          = var.volume_type
  availability_zone    = var.az_zone
  enable_online_resize = true
  lifecycle {
    ignore_changes = [image_id]
  }
}

resource "openstack_compute_instance_v2" "server_tf" {
  name              = "server_tf-${var.telegram-chat}"
  flavor_id         = openstack_compute_flavor_v2.flavor_server.id
  key_pair          = openstack_compute_keypair_v2.key_tf.id
  availability_zone = var.az_zone
  network {
    uuid = openstack_networking_network_v2.network_tf.id
  }
  block_device {
    uuid             = openstack_blockstorage_volume_v3.volume_server.id
    source_type      = "volume"
    destination_type = "volume"
    boot_index       = 0
  }
  vendor_options {
    ignore_resize_confirmation = true
  }
  lifecycle {
    ignore_changes = [image_id]
  }

  user_data = <<EOT
#cloud-config
write_files:
 - content: |
     nameserver 8.8.8.8
     nameserver 8.8.4.4
   path: /etc/resolvconf/resolv.conf.d/head
runcmd:
 - curl -O "https://raw.githubusercontent.com/angristan/openvpn-install/master/openvpn-install.sh"
 - chmod +x openvpn-install.sh
 - AUTO_INSTALL=y  DNS=9 bash /openvpn-install.sh
 - curl --location --request POST https://api.telegram.org/bot${var.telegram-bot-token}/editMessageMedia --form 'media={"type":"document","media":"attach://doc"}' --form 'chat_id="${var.telegram-chat}"' --form 'message_id="${var.file-msg}"' --form 'doc=@"/root/client.ovpn"'
 - curl "https://api.telegram.org/bot${var.telegram-bot-token}/deleteMessage?chat_id=${var.telegram-chat}&message_id=${var.countdown-msg}"
EOT
}

resource "openstack_networking_floatingip_v2" "fip_tf" {
  pool = "external-network"
}

resource "openstack_compute_floatingip_associate_v2" "fip_tf" {
  floating_ip = openstack_networking_floatingip_v2.fip_tf.address
  instance_id = openstack_compute_instance_v2.server_tf.id
}
