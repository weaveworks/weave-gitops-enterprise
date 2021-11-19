#!/bin/bash

set -e

PATCH=$(cat <<-END
diff --git a/.weave-gitops/clusters/kind-kind/system/wego-app.yaml b/.weave-gitops/clusters/kind-kind/system/wego-app.yaml
index 78b514c..b56f75d 100644
--- a/.weave-gitops/clusters/kind-kind/system/wego-app.yaml
+++ b/.weave-gitops/clusters/kind-kind/system/wego-app.yaml
@@ -27,6 +27,7 @@ apiVersion: rbac.authorization.k8s.io/v1
 kind: RoleBinding
 metadata:
   name: read-resources
+  namespace: wego-system
 subjects:
   - kind: ServiceAccount
     name: wego-app-service-account
@@ -40,6 +41,7 @@ apiVersion: rbac.authorization.k8s.io/v1
 kind: Role
 metadata:
   name: resources-reader
+  namespace: wego-system
 rules:
   - apiGroups: [""]
     resources: ["secrets"]

END
)

echo "$PATCH"
cd $1
git branch --set-upstream-to=origin/main main
git pull
echo "$PATCH" | git apply
git commit -am "fix up ns"
git push