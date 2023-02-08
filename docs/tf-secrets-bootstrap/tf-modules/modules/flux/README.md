# During development

oidc.eks.eu-north-1.amazonaws.com/id/F7D7CC46B978F2D9B83F4B3FD98257D3

- need to apply first namespace due to error

```bash
➜  flux git:(add-flux) ✗ tf apply -var "github_token=$GITHUB_TOKEN"                                                                                                    <aws:sts>
tls_private_key.main: Refreshing state... [id=d0ef7e8a82ad28f6064c8ee8b57f6e9bda6909d2]
kubernetes_namespace.flux_system: Refreshing state... [id=wge2205-leaf-flux-system]
╷
│ Error: Invalid for_each argument
│
│   on main.tf line 74, in resource "kubectl_manifest" "install":
│   74:   for_each   = { for v in local.install : lower(join("/", compact([v.data.apiVersion, v.data.kind, lookup(v.data.metadata, "namespace", ""), v.data.metadata.name]))) => v.content }
│     ├────────────────
│     │ local.install will be known only after apply
│
│ The "for_each" map includes keys derived from resource attributes that cannot be determined until apply, and so Terraform cannot determine the full set of keys that will
│ identify the instances of this resource.
│
│ When working with unknown values in for_each, it's better to define the map keys statically in your configuration and place apply-time results only in the map values.
│
│ Alternatively, you could use the -target planning option to first apply only the resources that the for_each value depends on, and then apply a second time to fully converge.
╵
╷
│ Error: Invalid for_each argument
│
│   on main.tf line 80, in resource "kubectl_manifest" "sync":
│   80:   for_each   = { for v in local.sync : lower(join("/", compact([v.data.apiVersion, v.data.kind, lookup(v.data.metadata, "namespace", ""), v.data.metadata.name]))) => v.content }
│     ├────────────────
│     │ local.sync will be known only after apply
│
│ The "for_each" map includes keys derived from resource attributes that cannot be determined until apply, and so Terraform cannot determine the full set of keys that will
│ identify the instances of this resource.
│
│ When working with unknown values in for_each, it's better to define the map keys statically in your configuration and place apply-time results only in the map values.
│
│ Alternatively, you could use the -target planning option to first apply only the resources that the for_each value depends on, and then apply a second time to fully converge.
╵

```

## when applied

```bash
➜  flux git:(add-flux) ✗ tf apply -var "github_token=$GITHUB_TOKEN" --target=kubernetes_namespace.flux_system                                                          <aws:sts>
kubernetes_namespace.flux_system: Refreshing state... [id=wge2205-leaf-flux-system]

No changes. Your infrastructure matches the configuration.
```
