FROM scratch
LABEL vendor="Dell Inc." \
      name="csm-metrics-powerstore" \
      summary="Dell Container Storage Modules (CSM) for Observability - Metrics for PowerStore" \
      description="Provides insight into storage usage and performance as it relates to the CSI (Container Storage Interface) Driver for Dell PowerStore" \
      version="2.0.0" \
      license="Apache-2.0"
ARG SERVICE
COPY $SERVICE/bin/service /service
ENTRYPOINT ["/service"]
