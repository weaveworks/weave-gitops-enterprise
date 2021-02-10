package payload

type ClusterInfo struct {
	// kube-system namespace UID
	ID string `json:"id"`
	// ProviderName assigned by the cloud provider
	Type  string     `json:"type"`
	Nodes []NodeInfo `json:"nodes"`
}

type NodeInfo struct {
	// .Status.NodeInfo.MachineID
	MachineID      string `json:"machineID"`
	Name           string `json:"name"`
	IsControlPlane bool   `json:"isControlPlane"`
	// .Status.NodeInfo.KubeletVersion
	KubeletVersion string `json:"kubeletVersion"`
}
