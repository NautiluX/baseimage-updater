package baseimg_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/NautiluX/baseimage-updater/pkg/baseimg"
	mock "github.com/NautiluX/baseimage-updater/pkg/baseimg/mocks"
	"github.com/golang/mock/gomock"
	_ "github.com/golang/mock/mockgen/model"
)

//go:generate mockgen -destination=mocks/mock_querier.go -package=mock -source=baseimg.go Querier
var _ = Describe("Baseimg", func() {
	var (
		mockCtrl    *gomock.Controller
		dockerfile  string
		updater     *BaseImageUpdater
		mockQuerier *mock.MockRegistryQuerier
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockQuerier = mock.NewMockRegistryQuerier(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	JustBeforeEach(func() {
		var err error
		updater, err = NewBaseImageUpdater(dockerfile, "^[0-9]+\\.[0-9]+-[0-9]+$")
		Expect(err).NotTo(HaveOccurred())

		updater.RegistryQuerier = mockQuerier
	})

	Context("Single FROM line", func() {
		var (
			tagList []string
		)
		BeforeEach(func() {
			dockerfile = "FROM registry.access.redhat.com/ubi8/ubi-micro:8.4-0"
		})
		JustBeforeEach(func() {
			mockQuerier.EXPECT().GetTag(gomock.Any()).Return("8.4-0")
			mockQuerier.EXPECT().GetName(gomock.Any()).Return("registry.access.redhat.com/ubi8/ubi-micro:8.4-0").Times(2)
			mockQuerier.EXPECT().ListTags(gomock.Any()).Return(tagList, nil)
			mockQuerier.EXPECT().GetFullTag(gomock.Any(), gomock.Any()).DoAndReturn(func(input, tag string) string {
				return "registry.access.redhat.com/ubi8/ubi-micro:" + tag
			})
		})
		Context("When 2 higher versions are returned", func() {
			BeforeEach(func() {
				tagList = []string{"8.5-999", "8.4-1"}
			})
			It("Updates the ubi-micro base image to the highest version", func() {
				newDockerfile, err := updater.UpdateBaseImages()
				Expect(err).NotTo(HaveOccurred())
				Expect(newDockerfile).To(Equal("FROM registry.access.redhat.com/ubi8/ubi-micro:8.5-999"))
			})
		})
		Context("When some base image tag doesn't match the base image regex but is higher version", func() {
			BeforeEach(func() {
				tagList = []string{"8.5-999-source", "8.4-1"}
			})
			It("Updates the ubi-micro base image to the matching version", func() {
				newDockerfile, err := updater.UpdateBaseImages()
				Expect(err).NotTo(HaveOccurred())
				Expect(newDockerfile).To(Equal("FROM registry.access.redhat.com/ubi8/ubi-micro:8.4-1"))
			})

		})
		Context("When some tags have a smaller minor version number", func() {
			BeforeEach(func() {
				tagList = []string{"8.3-999", "8.4-1"}
			})
			It("Updates the ubi-micro base image to the matching version", func() {
				newDockerfile, err := updater.UpdateBaseImages()
				Expect(err).NotTo(HaveOccurred())
				Expect(newDockerfile).To(Equal("FROM registry.access.redhat.com/ubi8/ubi-micro:8.4-1"))
			})

		})
		Context("When some tags have a greater prerelease version number", func() {
			BeforeEach(func() {
				tagList = []string{"8.4-2", "8.4-1"}
			})
			It("Updates the ubi-micro base image to the matching version", func() {
				newDockerfile, err := updater.UpdateBaseImages()
				Expect(err).NotTo(HaveOccurred())
				Expect(newDockerfile).To(Equal("FROM registry.access.redhat.com/ubi8/ubi-micro:8.4-2"))
			})

		})
		Context("When some tags have a non-numeric prerelease version number", func() {
			BeforeEach(func() {
				tagList = []string{"8.5-asdf", "8.4-1"}
			})
			It("Updates the ubi-micro base image to the matching version", func() {
				newDockerfile, err := updater.UpdateBaseImages()
				Expect(err).NotTo(HaveOccurred())
				Expect(newDockerfile).To(Equal("FROM registry.access.redhat.com/ubi8/ubi-micro:8.4-1"))
			})

		})
	})
	Context("When multiple FROM commands exist in the dockerfile", func() {
		Context("When one FROM command doesn't use the matching regex for versioning", func() {
			BeforeEach(func() {
				dockerfile = "FROM quay.io/app-sre/boilerplate:image-v2.3.2\nFROM registry.access.redhat.com/ubi8/ubi-micro:8.4-0"
				mockQuerier.EXPECT().GetTag(gomock.Any()).Return("v2.3.2")
				mockQuerier.EXPECT().GetTag(gomock.Any()).Return("8.4-0")
				mockQuerier.EXPECT().GetName(gomock.Any()).Return("registry.access.redhat.com/ubi8/ubi-micro:8.4-0").Times(2)
				mockQuerier.EXPECT().ListTags(gomock.Any()).Return([]string{"8.5-999", "8.4-1"}, nil)
				mockQuerier.EXPECT().GetFullTag(gomock.Any(), gomock.Any()).Return("registry.access.redhat.com/ubi8/ubi-micro:8.5-999")
			})
			It("Updates only the ubi-micro base image", func() {
				newDockerfile, err := updater.UpdateBaseImages()
				Expect(err).NotTo(HaveOccurred())
				Expect(newDockerfile).To(Equal("FROM quay.io/app-sre/boilerplate:image-v2.3.2\nFROM registry.access.redhat.com/ubi8/ubi-micro:8.5-999"))
			})
		})

	})
})
