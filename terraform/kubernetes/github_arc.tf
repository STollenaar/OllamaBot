resource "helm_release" "github_arc_runner_ollamabot" {
  name       = "arc-runner-ollamabot"
  namespace  = data.terraform_remote_state.kubernetes_cluster.outputs.github_arc.namespace.metadata.0.name
  repository = "oci://ghcr.io/actions/actions-runner-controller-charts"
  chart      = "gha-runner-scale-set"
  version    = data.terraform_remote_state.kubernetes_cluster.outputs.github_arc.version
  values = [templatefile("${path.module}/conf/arc-runner-values.yaml", {
    github_config_url = "https://github.com/STollenaar/ollamabot"
    github_secret     = data.terraform_remote_state.kubernetes_cluster.outputs.github_arc.secret_name
  })]
}

resource "kubernetes_role_binding" "github_arc_admin" {
  metadata {
    name      = "namespace-admin-binding"
    namespace = kubernetes_namespace_v1.ollamabot.id
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "admin"
  }

  subject {
    kind      = "ServiceAccount"
    name      = "arc-runner-ollamabot-gha-rs-kube-mode"
    namespace = data.terraform_remote_state.kubernetes_cluster.outputs.github_arc.namespace.metadata.0.name
  }
}
