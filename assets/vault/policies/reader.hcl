# `*` is a glob only at the end of a path: `github-*` covers every github-<repo> mount.
path "github-*" {
  capabilities = ["read", "list"]
}
path "kubernetes-*" {
  capabilities = ["read", "list"]
}
# developer machines (Pulumi tokens, per-host GCP keys), created by github-infrastructure
path "dev-infrastructure/*" {
  capabilities = ["read", "list"]
}
