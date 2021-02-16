package payload

import (
	v1 "k8s.io/api/core/v1"
)

type KubernetesEvent struct {
	Token string   `json:"token"`
	Event v1.Event `json:"event"`
}
