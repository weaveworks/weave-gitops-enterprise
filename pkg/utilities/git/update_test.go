package git

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
)

var crbPrev = `kind: ClusterRoleBinding
metadata:
  name: foo
roleRef:
  # A comment
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: wkp-flux
subjects:
  - kind: ServiceAccount
    name: flux
    namespace: wkp-flux
`

var crbNext = `kind: ClusterRoleBinding
metadata:
  name: bar
roleRef:
  # A comment
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: wkp-flux
subjects:
  - kind: ServiceAccount
    name: flux
    namespace: wkp-flux
`

var elaboratePrev = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: wks-controller
  namespace: system
  labels:
    name: wks-controller
    control-plane: wks-controller
    controller-tools.k8s.io: "1.0"
spec:
  replicas: 1
  selector:
    matchLabels:
      name: wks-controller
  template:
    metadata:
      labels:
        name: wks-controller
        control-plane: wks-controller
        controller-tools.k8s.io: "1.0"
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
      # Allow scheduling on master nodes. This is required because during
      # bootstrapping of the cluster, we may initially have just one master,
      # and would then need to deploy this controller there to set the entire
      # cluster up.
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
          operator: Exists
        - # Mark this as a critical addon:
          key: CriticalAddonsOnly
          operator: Exists
        - # Only schedule on nodes which are ready and reachable:
          effect: NoExecute
          key: node.alpha.kubernetes.io/notReady
          operator: Exists
        - effect: NoExecute
          key: node.alpha.kubernetes.io/unreachable
          operator: Exists
      containers:
        - name: controller
          image: docker.io/jrryjcksn/wks-controller:2019-09-18-scale-footloose-8e831136-WIP
          env:
            - name: BRIDGE_ADDRESS
            - value: 172.17.0.1
          command:
            - /bin/controller
            - --verbose
          resources:
            limits:
              cpu: 100m
              memory: 30Mi
            requests:
              cpu: 100m
              memory: 20Mi
      imagePullSecrets:
        - name: wks-image-secret
`

var elaborateNext = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: wks-controller
  namespace: system
  labels:
    name: wks-controller
    control-plane: wks-controller
    controller-tools.k8s.io: "1.0"
spec:
  replicas: 1
  selector:
    matchLabels:
      name: wks-controller
  template:
    metadata:
      labels:
        name: wks-controller
        control-plane: wks-controller
        controller-tools.k8s.io: "1.0"
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
        # Allow scheduling on master nodes. This is required because during
        # bootstrapping of the cluster, we may initially have just one master,
        # and would then need to deploy this controller there to set the entire
        # cluster up.
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
          operator: Exists
        - # Mark this as a critical addon:
          key: CriticalAddonsOnly
          operator: Exists
        - # Only schedule on nodes which are ready and reachable:
          effect: NoExecute
          key: node.alpha.kubernetes.io/notReady
          operator: Exists
        - effect: NoExecute
          key: node.alpha.kubernetes.io/unreachable
          operator: Exists
      containers:
        - name: controller
          image: docker.io/jrryjcksn/wks-controller:2019-09-18-scale-footloose-8e831136-WIP
          env:
            - name: BRIDGE_ADDRESS
            - value: 172.17.0.1
          command:
            - /bin/controller
            - --verbose
          resources:
            limits:
              cpu: 200m
              memory: 30Mi
            requests:
              cpu: 100m
              memory: 20Mi
      imagePullSecrets:
        - name: wks-image-secret
`

type testdata struct {
	kind, namespace, name, before, after, path, value string
}

var tests = []testdata{
	{"ClusterRoleBinding", "default", "foo", crbPrev, crbNext, "metadata.name", "bar"},
	{"Deployment", "system", "wks-controller", elaboratePrev, elaborateNext, "spec.template.spec.containers.0.resources.limits.cpu", "200m"},
}

func TestUpdate(t *testing.T) {
	for _, data := range tests {
		var input yaml.Node
		err := yaml.Unmarshal([]byte(data.before), &input)
		assert.NoError(t, err)
		objectNode := findObjectNode(&input, data.kind, data.namespace, data.name)
		assert.NotNil(t, objectNode)
		path := strings.Split(data.path, ".")
		err = UpdateNestedFields(objectNode, data.value, path...)
		assert.NoError(t, err)
		var output bytes.Buffer
		encoder := yaml.NewEncoder(&output)
		defer encoder.Close()
		encoder.SetIndent(2)
		err = encoder.Encode(&input)
		assert.NoError(t, err)
		assert.Equal(t, data.after, output.String())
	}
}
