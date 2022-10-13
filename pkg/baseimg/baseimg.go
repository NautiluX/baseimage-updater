package baseimg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	registry "github.com/NautiluX/baseimage-updater/pkg/registry"
	log "github.com/sirupsen/logrus"

	"github.com/Masterminds/semver"
	"github.com/asottile/dockerfile"
)

type BaseImageUpdater struct {
	Dockerfile string
	TagRegex   string
	tagRe      *regexp.Regexp
}

func NewBaseImageUpdater(dockerfile string, tagRegex string) (*BaseImageUpdater, error) {

	tagRe, err := regexp.Compile(tagRegex)
	if err != nil {
		return &BaseImageUpdater{}, fmt.Errorf("can't parse tag regex: %v", err)
	}

	return &BaseImageUpdater{dockerfile, tagRegex, tagRe}, nil

}

func (b *BaseImageUpdater) UpdateBaseImages() (output string, err error) {
	output = b.Dockerfile
	myReader := strings.NewReader(b.Dockerfile)
	commands, err := dockerfile.ParseReader(myReader)
	if err != nil {
		return output, fmt.Errorf("can't read input string: %v", err)
	}
	for _, cmd := range commands {
		newCommand, err := b.processDockerfileCommand(cmd)
		if err != nil {
			return output, fmt.Errorf("failed to parse command %s: %v", output, err)
		}
		if newCommand == cmd.Original {
			continue
		}
		log.Infof("Old FROM line: %s\n", cmd.Original)
		log.Infof("New FROM line: %s\n", newCommand)
		output = strings.ReplaceAll(output, cmd.Original, newCommand)
	}
	err = nil
	return
}

func (b *BaseImageUpdater) processDockerfileCommand(cmd dockerfile.Command) (string, error) {
	if strings.ToUpper(cmd.Cmd) != "FROM" {
		return cmd.Original, nil
	}
	log.Trace(cmd.Cmd)
	log.Trace(len(cmd.Value))
	for _, value := range cmd.Value {
		if strings.ToUpper(value) == "AS" {
			break
		}

		registryQuerier, err := registry.NewQuerier(value)
		if err != nil {
			log.Infof("failed to initialize registry querier: for value %s: %v. Ignoring", value, err)
			continue
		}
		newestVersion := registryQuerier.GetTag()

		tags, err := registryQuerier.ListTags()
		if err != nil {
			log.Infof("failed to query registry %s for tags: %v. Ignoring", value, err)
			continue
		}

		for _, tag := range tags {
			validNewerVersion, err := b.isNewerVersion(newestVersion, tag)
			if err != nil {
				return cmd.Original, fmt.Errorf("failed to parse version %s: %v", tag, err)
			}
			if validNewerVersion {
				newestVersion = tag
			}
		}
		log.Infof("Newest version found: %s\n", newestVersion)
		newTag := registryQuerier.GetFullTag(newestVersion)
		log.Infof("Newest tag found: %s\n", newTag)
		if newTag == registryQuerier.GetName() {
			log.Info("No version change, not touching FROM line.")
			continue
		}
		newFrom := strings.ReplaceAll(cmd.Original, registryQuerier.GetName(), newTag)
		return newFrom, nil
	}
	return cmd.Original, nil
}

func (b *BaseImageUpdater) isNewerVersion(newestVersion, tag string) (bool, error) {
	if !b.tagRe.Match([]byte(tag)) {
		log.Tracef("Tag %s doesn't match tag regex %s. Ignoring.\n", tag, b.TagRegex)
		return false, nil
	}
	log.Trace(tag)
	verCurrent, err := semver.NewVersion(tag)
	if err != nil {
		log.Tracef("Not a valid SemVer %s, ignoring.\n", tag)
		return false, nil
	}
	verNewest, err := semver.NewVersion(newestVersion)
	if err != nil {
		log.Tracef("Not a valid SemVer %s, ignoring.\n", tag)
		return false, nil
	}
	cGreater, err := semver.NewConstraint(">" + newestVersion)
	if err != nil {
		log.Infof("Couldn't parse constraint with version %s\n", newestVersion)
		return false, nil
	}

	if verCurrent.Prerelease() == "" && verNewest.Prerelease() != "" {
		log.Tracef("%s doesn't contain a Prerelease, but input does", tag)
		return false, nil
	}
	if cGreater.Check(verCurrent) {
		log.Tracef("Greater version found: %s > %s\n", tag, newestVersion)
		newestVersion = tag
		return true, nil
	}
	cGreaterEqual, err := semver.NewConstraint(">=" + newestVersion)
	if err != nil {
		log.Tracef("Couldn't parse constraint with version %s\n", newestVersion)
		return false, nil
	}
	if cGreaterEqual.Check(verCurrent) {
		prereleaseIntCurrent, err := strconv.Atoi(verCurrent.Prerelease())
		if err != nil {
			log.Tracef("not using prerelease string for comparison, not an int: %s\n", verCurrent.Prerelease())
			return false, nil
		}
		prereleaseIntNewest, err := strconv.Atoi(verNewest.Prerelease())
		if err != nil {
			log.Tracef("not using prerelease string for comparison, not an int: %s\n", verNewest.Prerelease())
			return false, nil
		}
		if prereleaseIntCurrent > prereleaseIntNewest {
			log.Tracef("Greater prerelase found: %s > %s\n", tag, newestVersion)
			return true, nil
		}

	}
	return false, nil
}
