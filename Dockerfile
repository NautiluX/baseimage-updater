FROM quay.io/app-sre/boilerplate:image-v2.3.2 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# Copy the go source
COPY . .

# Build
RUN make build

FROM registry.access.redhat.com/ubi8/ubi-micro:8.6-484
WORKDIR /
COPY --from=builder /workspace/baseimage-updater /baseimage-updater

ENTRYPOINT ["/baseimage-updater"]
