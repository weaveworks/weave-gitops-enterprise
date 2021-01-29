module github.com/weaveworks/wks/cmd/event-writer

go 1.13

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.3.1
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/google/uuid v1.1.2
	github.com/nats-io/nats-server/v2 v2.1.7
	github.com/nats-io/nats.go v1.10.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	github.com/tj/assert v0.0.3
	gorm.io/datatypes v1.0.0
	gorm.io/driver/sqlite v1.1.3
	gorm.io/gorm v1.20.12
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.20.2
