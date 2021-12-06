module github.com/dell/csm-metrics-powerstore

go 1.14

require (
	github.com/dell/gocsi v1.3.1
	github.com/dell/gopowerstore v1.5.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/mock v1.4.3
	github.com/jarcoal/httpmock v1.0.6 // indirect
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.4 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v0.7.0
	go.opentelemetry.io/otel/exporters/otlp v0.7.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.7.0
	golang.org/x/crypto v0.0.0-20201216223049-8b5274cf687f // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
)
