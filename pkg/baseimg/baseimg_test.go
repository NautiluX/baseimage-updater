package baseimg_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/NautiluX/baseimage-updater/pkg/baseimg"
)

var _ = Describe("Baseimg", func() {
	Context("Exmaple Dockerfile with ubi-micro image", func() {
		var (
			dockerfile string
		)
		BeforeEach(func() {
			dockerfile = rmoDockerfile("registry.access.redhat.com/ubi8/ubi-micro:8.4-0")
		})
		It("Updates the ubi-micro base image", func() {
			updater, err := NewBaseImageUpdater(dockerfile, "^[0-9]+\\.[0-9]+-[0-9]+$")
			Expect(err).NotTo(HaveOccurred())
			newDockerfile, err := updater.UpdateBaseImages()
			Expect(err).NotTo(HaveOccurred())
			Expect(newDockerfile).NotTo(Equal(rmoDockerfile("registry.access.redhat.com/ubi8/ubi-micro:8.4-0")))
		})
	})
	Context("When multiple FROM commands exist in the dockerfile", func() {})
	Context("When some base image tag doesn't match the base image regex", func() {})
	Context("When some tags don't match the base image regex", func() {})
	Context("When some tags have a greater minor version number", func() {})
	Context("When some tags have a greater prerelease version number", func() {})
	Context("When some tags have a non-numeric prerelease version number", func() {})
})

func rmoDockerfile(image string) string {
	return `FROM quay.io/app-sre/boilerplate:image-v2.3.2 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# Copy the go source
COPY . .

# Build
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod vendor -o manager main.go
RUN make go-build

FROM ` + image + `
WORKDIR /
COPY --from=builder /workspace/build/_output/bin/* /manager
USER nonroot:nonroot

ENTRYPOINT ["/manager"] `

}
