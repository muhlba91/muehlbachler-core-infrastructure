path "secret/*" {
  capabilities = ["read", "list"]
}
path "github-*/*" {
  capabilities = ["read", "list"]
}
path "kubernetes-*/*" {
  capabilities = ["read", "list"]
}
