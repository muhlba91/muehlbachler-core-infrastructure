path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
# `*` is a glob only at the end of a path: `github-*` covers every github-<repo> mount.
path "github-*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "kubernetes-*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
