package payload

type ClusterInfo struct {
	Token   string  `json:"token"`
	Cluster Cluster `json:"cluster"`
}

type Cluster struct {
	// kube-system namespace UID
	ID string `json:"id"`
	// ProviderName assigned by the cloud provider
	Type  string `json:"type"`
	Nodes []Node `json:"nodes"`
}

type Node struct {
	// .Status.NodeInfo.MachineID
	MachineID      string `json:"machineID"`
	Name           string `json:"name"`
	IsControlPlane bool   `json:"isControlPlane"`
	// .Status.NodeInfo.KubeletVersion
	KubeletVersion string `json:"kubeletVersion"`
}
