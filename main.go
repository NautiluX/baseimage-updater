package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/NautiluX/baseimage-updater/pkg/baseimg"
	"github.com/kylelemons/godebug/diff"
)

func main() {

	dockerfileContent := "FROM registry.access.redhat.com/ubi8/ubi-micro:8.5-836 as BASE"
	tagRegex := "^[0-9]+\\.[0-9]+-[0-9]+$"
	filename := ""
	if len(os.Args) >= 2 {
		filename = os.Args[1]
		bytes, err := os.ReadFile(filename)
		if err != nil {
			log.Errorf("Can't read file " + filename)
			return
		}
		dockerfileContent = string(bytes)
	}
	if len(os.Args) == 3 {
		tagRegex = os.Args[2]
	}
	log.SetLevel(log.TraceLevel)
	updater, err := baseimg.NewBaseImageUpdater(dockerfileContent, tagRegex)
	if err != nil {
		log.Errorf("Failed to initialize base image updater in dockerfile: %v\n", err)
	}
	newDockerfile, err := updater.UpdateBaseImages()
	if err != nil {
		log.Errorf("Failed to update base images in dockerfile: %v\n", err)
	}
	if newDockerfile == dockerfileContent {
		log.Infof("No new base images found.")
		return
	}
	log.Infof("Diff:\n%s\n", diff.Diff(dockerfileContent, newDockerfile))
	log.Infof("New Dockerfile:\n%s\n", newDockerfile)
	if filename != "" {
		info, err := os.Stat(filename)
		if err != nil {
			log.Errorf("Couldn't get file permissions: %v", err)
		}
		os.WriteFile(filename, []byte(newDockerfile), info.Mode())
	}
}
