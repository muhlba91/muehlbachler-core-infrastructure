path "secret/*" {
  capabilities = ["read", "list"]
}
# `*` is a glob only at the end of a path: `github-*` covers every github-<repo> mount.
path "github-*" {
  capabilities = ["read", "list"]
}
path "kubernetes-*" {
  capabilities = ["read", "list"]
}
