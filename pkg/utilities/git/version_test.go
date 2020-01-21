package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var machinesPrevNoVersion = `apiVersion: v1
items:
- apiVersion: cluster.k8s.io/v1alpha1
  kind: Machine
  metadata:
    labels:
      set: master
    name: master-0
    namespace: weavek8sops
  spec:
    versions:
      kubelet: 1.14.1
    providerSpec:
      value:
        apiVersion: baremetalproviderspec/v1alpha1
        kind: BareMetalMachineProviderSpec
        private:
          address: 172.17.0.2
          port: 22
        public:
          address: 127.0.0.1
          port: 2222
- apiVersion: cluster.k8s.io/v1alpha1
  kind: Machine
  metadata:
    labels:
      set: worker
    name: worker-0
    namespace: weavek8sops
  spec:
    providerSpec:
      value:
        apiVersion: baremetalproviderspec/v1alpha1
        kind: BareMetalMachineProviderSpec
        private:
          address: 172.17.0.3
          port: 22
        public:
          address: 127.0.0.1
          port: 2223
kind: List
`

var machinesPrev = `apiVersion: v1
items:
- apiVersion: cluster.k8s.io/v1alpha1
  kind: Machine
  metadata:
    labels:
      set: master
    name: master-0
    namespace: weavek8sops
  spec:
    versions:
      kubelet: 1.14.1
    providerSpec:
      value:
        apiVersion: baremetalproviderspec/v1alpha1
        kind: BareMetalMachineProviderSpec
        private:
          address: 172.17.0.2
          port: 22
        public:
          address: 127.0.0.1
          port: 2222
- apiVersion: cluster.k8s.io/v1alpha1
  kind: Machine
  metadata:
    labels:
      set: worker
    name: worker-0
    namespace: weavek8sops
  spec:
    versions:
      kubelet: 1.14.10
    providerSpec:
      value:
        apiVersion: baremetalproviderspec/v1alpha1
        kind: BareMetalMachineProviderSpec
        private:
          address: 172.17.0.3
          port: 22
        public:
          address: 127.0.0.1
          port: 2223
kind: List
`

var machinesNext = `apiVersion: v1
items:
- apiVersion: cluster.k8s.io/v1alpha1
  kind: Machine
  metadata:
    labels:
      set: master
    name: master-0
    namespace: weavek8sops
  spec:
    versions:
      kubelet: 1.15.7
    providerSpec:
      value:
        apiVersion: baremetalproviderspec/v1alpha1
        kind: BareMetalMachineProviderSpec
        private:
          address: 172.17.0.2
          port: 22
        public:
          address: 127.0.0.1
          port: 2222
- apiVersion: cluster.k8s.io/v1alpha1
  kind: Machine
  metadata:
    labels:
      set: worker
    name: worker-0
    namespace: weavek8sops
  spec:
    versions:
      kubelet: 1.15.7
    providerSpec:
      value:
        apiVersion: baremetalproviderspec/v1alpha1
        kind: BareMetalMachineProviderSpec
        private:
          address: 172.17.0.3
          port: 22
        public:
          address: 127.0.0.1
          port: 2223
kind: List
`

func writeMachinesConfig(data string) (string, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "machines-*.yaml")
	if err != nil {
		return "", err
	}
	_, err = tmpFile.Write([]byte(data))
	if err != nil {
		return "", err
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}

func readMachinesConfig(fileName string) (string, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func TestGetMachinesK8sVersions_Standard(t *testing.T) {
	fileName, err := writeMachinesConfig(machinesPrev)
	assert.NoError(t, err)
	defer os.Remove(fileName)

	versions, err := GetMachinesK8sVersions("", fileName)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1.14.1", "1.14.10"}, versions)
}

func TestGetMachinesK8sVersions_NoVersion(t *testing.T) {
	fileName, err := writeMachinesConfig(machinesPrevNoVersion)
	assert.NoError(t, err)
	defer os.Remove(fileName)

	_, err = GetMachinesK8sVersions("", fileName)
	assert.Equal(t, fmt.Errorf("Kubelet version missing for a node in %s", fileName), err)
}

func TestUpdateMachinesK8sVersions_Standard(t *testing.T) {
	fileName, err := writeMachinesConfig(machinesPrev)
	assert.NoError(t, err)
	defer os.Remove(fileName)

	err = UpdateMachinesK8sVersions("", fileName, "1.15.7")
	assert.NoError(t, err)

	content, err := readMachinesConfig(fileName)
	assert.NoError(t, err)
	assert.Equal(t, content, machinesNext)
}
