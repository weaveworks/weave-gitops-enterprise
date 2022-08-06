# Updating the CRDs

After updating the files in api/capi, api/templates or api/gitopstemplate you **MUST** rebuild the CRDs and generated code.

```shell
$ make crd-generate crd-manifests
```
