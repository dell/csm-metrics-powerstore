ARG BASEIMAGE
ARG GOIMAGE

# Build the sdk binary
FROM $GOIMAGE as builder

# Set envirment variable
ENV APP_NAME csm-metrics-powerstore
ENV CMD_PATH cmd/metrics-powerstore/main.go

# Copy application data into image
COPY . /go/src/$APP_NAME
WORKDIR /go/src/$APP_NAME

# Build the binary
RUN go install github.com/golang/mock/mockgen@v1.6.0
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/src/service /go/src/$APP_NAME/$CMD_PATH

# Build the sdk image
FROM $BASEIMAGE as final
LABEL vendor="Dell Technologies" \
      maintainer="Dell Technologies" \
      name="csm-metrics-powerstore" \
      summary="Dell Container Storage Modules (CSM) for Observability - Metrics for PowerStore" \
      description="Provides insight into storage usage and performance as it relates to the CSI (Container Storage Interface) Driver for Dell PowerStore" \
      release="1.14.0" \
      version="1.12.0" \
      license="Apache-2.0"
COPY --from=builder /go/src/service /
ENTRYPOINT ["/service"]
