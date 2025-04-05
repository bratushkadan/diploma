locals {
  opensearch_creds = yamldecode(file("${path.module}/opensearch-creds.yaml"))
}

# data "yandex_compute_image" "container_optimized_image" {
#   family = "container-optimized-image"
# }

locals {
  # coi_id = yandex_compute_image.container_optimized_image
  coi_id = "fd8r54pj9a0ic0ftsbvf"
}

resource "yandex_compute_disk" "opensearch_data" {
  name = "opensearch-data"
  type = "network-ssd"
  zone = "ru-central1-b"
  size = 10
}

resource "yandex_compute_snapshot_schedule" "opensearch_data" {
  name = "opensearch-data"

  schedule_policy {
    expression = "0 0 * * *"
  }

  snapshot_count = 1

  snapshot_spec {
    description = "opensearch data disk snapshot"
  }

  disk_ids = [yandex_compute_disk.opensearch_data.id]
}

resource "yandex_compute_instance" "opensearch" {
  name = "opensearch"
  boot_disk {
    initialize_params {
      image_id = local.coi_id
    }
  }
  zone = "ru-central1-b"
  network_interface {
    subnet_id          = yandex_vpc_subnet.opensearch.id
    nat                = true
    security_group_ids = [yandex_vpc_security_group.this.id]
  }
  resources {
    cores  = 2
    memory = 2
  }
  secondary_disk {
    disk_id     = yandex_compute_disk.opensearch_data.id
    device_name = local.opensearch_instance_data_disk_device_name
  }
  metadata = {
    docker-compose = templatefile("${path.module}/config/opensearch-containers-docker-compose.yaml", {
      opensearch_password = local.opensearch_creds.password,
    })
    user-data = templatefile("${path.module}/config/opensearch-vm-cloud-config.yaml", {
      data_disk_device_name = local.opensearch_instance_data_disk_device_name
    })
  }
}
