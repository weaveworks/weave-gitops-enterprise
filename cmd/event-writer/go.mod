module github.com/weaveworks/wks/cmd/event-writer

go 1.13

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.3.1
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/go-openapi/strfmt v0.19.5
	github.com/google/uuid v1.2.0
	github.com/nats-io/nats-server/v2 v2.1.7
	github.com/nats-io/nats.go v1.10.0
	github.com/prometheus/alertmanager v0.21.0
	github.com/prometheus/common v0.10.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	github.com/weaveworks/wks/common v0.0.0
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c // indirect
	gorm.io/datatypes v1.0.0
	gorm.io/gorm v1.20.12
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.20.2

replace github.com/weaveworks/wks/common => ../../common
