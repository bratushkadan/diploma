resource "yandex_vpc_network" "this" {
  folder_id = local.folder_id

  name        = "e-com"
  description = "network for e-com"
}

resource "yandex_vpc_subnet" "opensearch" {
  v4_cidr_blocks = [local.opensearch_subnet_cidr]
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.this.id
}

resource "yandex_vpc_security_group" "this" {
  folder_id = local.folder_id

  name        = "e-com-sg"
  description = "e-com security group"

  network_id = yandex_vpc_network.this.id
}

resource "yandex_vpc_security_group_rule" "opensearch_ssh_ingress" {
  security_group_binding = yandex_vpc_security_group.this.id
  direction              = "ingress"
  v4_cidr_blocks         = ["0.0.0.0/0"]
  port                   = 22
  protocol               = "TCP"
}
resource "yandex_vpc_security_group_rule" "opensearch_egress" {
  security_group_binding = yandex_vpc_security_group.this.id
  direction              = "egress"
  description            = "permit any"
  v4_cidr_blocks         = ["0.0.0.0/0"]
  protocol               = "ANY"
}
