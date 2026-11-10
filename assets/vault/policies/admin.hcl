path "sys/health" {
  capabilities = ["read", "sudo"]
}
path "sys/policies/acl" {
  capabilities = ["list"]
}
path "sys/policies/acl/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "sys/auth/*" {
  capabilities = ["create", "update", "delete", "sudo"]
}
path "sys/auth" {
  capabilities = ["read"]
}
# `*` is a glob only at the end of a path: `github-*` covers every github-<repo> mount.
path "github-*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "kubernetes-*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
# developer machines (Pulumi tokens, per-host GCP keys), created by github-infrastructure
path "dev-infrastructure/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "sys/mounts" {
  capabilities = ["read"]
}
