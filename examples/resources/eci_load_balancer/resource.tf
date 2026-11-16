data "eci_pricing" "lb_pricing" {
  name         = "Load Balancer"
  pricing_type = "ondemand"
}

resource "eci_load_balancer" "my_load_balancer" {
  name                = "web"
  attached_network_id = "0f4c2a3e-2a31-4b8e-9a51-8d1f8c0c6b10"
  public_ip_id        = "6b1d9e57-3c2f-4f0a-8e4e-1a7b2c3d4e5f"
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
