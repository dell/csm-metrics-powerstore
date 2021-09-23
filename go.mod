module github.com/dell/csm-metrics-powerstore

go 1.14

require (
	github.com/dell/csi-powerstore v0.0.0-20210923102620-2b6d0e23ab2e
	github.com/dell/gofsutil v1.6.0
	github.com/dell/gopowerstore v1.5.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/mock v1.4.3
	github.com/jarcoal/httpmock v1.0.6 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v0.7.0
	go.opentelemetry.io/otel/exporters/otlp v0.7.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.7.0
	golang.org/x/crypto v0.0.0-20201216223049-8b5274cf687f // indirect
	google.golang.org/grpc v1.38.0
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
)
