package payload

type CAPIClusterInfo struct {
	Token        string        `json:"token"`
	CAPIClusters []CAPICluster `json:"capiClusters"`
}

type CAPICluster struct {
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	CAPIVersion   string `json:"capiVersion"`
	EncodedObject string `json:"object"`
}
