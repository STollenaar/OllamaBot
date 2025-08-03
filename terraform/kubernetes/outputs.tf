output "namespace" {
  value = kubernetes_namespace.ollamabot
}

output "external_secret" {
  value = kubernetes_manifest.external_secret.manifest
}