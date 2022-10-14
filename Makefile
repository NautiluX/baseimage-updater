IMG?=quay.io/mdewald/baseimage-updater

.PHONY:
generate:
	go generate ./...

container-build:
	podman build . -t $(IMG)

container-push: container-build
	podman push $(IMG)

docker-build: container-build
docker-push: container-push

test:
	go test ./...

build:
	go build -mod=mod
