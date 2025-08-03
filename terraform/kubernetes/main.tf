locals {
  name = "ollamabot"
}

resource "kubernetes_namespace" "ollamabot" {
  metadata {
    name = local.name
  }
}

resource "kubernetes_service_v1" "ollamabot" {
  metadata {
    name      = "ollamabot"
    namespace = kubernetes_namespace.ollamabot.metadata.0.name
  }
  spec {
    selector = {
      "app" = local.name
    }
    port {
      name        = "router"
      target_port = 8080
      port        = 80
    }
  }
}