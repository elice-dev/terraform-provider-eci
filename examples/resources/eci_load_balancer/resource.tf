data "eci_pricing" "lb_pricing" {
  name         = "Load Balancer"
  pricing_type = "ondemand"
}

data "eci_pricing" "ip_pricing" {
  name         = "Public IP"
  pricing_type = "ondemand"
}

# A load balancer needs a public ip that is not attached to a network interface.
resource "eci_public_ip" "lb_ip" {
  dr         = false
  pricing_id = data.eci_pricing.ip_pricing.id
  tags       = {}
}

resource "eci_load_balancer" "my_load_balancer" {
  name                = "web"
  attached_network_id = "0f4c2a3e-2a31-4b8e-9a51-8d1f8c0c6b10"
  public_ip_id        = eci_public_ip.lb_ip.id
  pricing_id          = data.eci_pricing.lb_pricing.id

  rules = [
    {
      proto     = "TCP"
      port      = 80
      algorithm = "roundrobin"
      servers = [
        {
          ip                   = "10.0.0.11"
          network_interface_id = "4adf2682-d8f3-451c-8bfc-3383deb424a5"
          port                 = 8080
        },
      ]
      healthcheck = {
        proto               = "HTTP"
        port                = 8080
        path                = "/healthz"
        interval            = 5
        timeout             = 3
        healthy_threshold   = 2
        unhealthy_threshold = 3
      }
    },
  ]

  tags = {
    "created-by" : "terraform"
  }
}
