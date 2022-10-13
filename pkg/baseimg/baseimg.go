package baseimg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Masterminds/semver"
	"github.com/asottile/dockerfile"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func UpdateBaseImages(input string, tagRegex string) (output string, err error) {
	output = input
	myReader := strings.NewReader(input)
	commands, err := dockerfile.ParseReader(myReader)
	if err != nil {
		return input, fmt.Errorf("can't read input string: %v", err)
	}
	for _, cmd := range commands {
		newCommand, err := processDockerfileCommand(cmd, tagRegex)
		if err != nil {
			return input, fmt.Errorf("failed to parse command %s: %v", input, err)
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

func processDockerfileCommand(cmd dockerfile.Command, tagRegex string) (string, error) {
	if strings.ToUpper(cmd.Cmd) != "FROM" {
		return cmd.Original, nil
	}
	log.Trace(cmd.Cmd)
	log.Trace(len(cmd.Value))
	for _, value := range cmd.Value {
		if strings.ToUpper(value) == "AS" {
			break
		}
		ref, err := name.ParseReference(value)
		if err != nil {
			log.Infof("Couldn't parse image %s, ignoring.\n", value)
			continue
		}

		log.Infof("Identifier: %s\n", ref.Identifier())
		log.Infof("Name: %s\n", ref.Name())
		log.Infof("RegistryStr: %s\n", ref.Context().RegistryStr())
		log.Infof("RepositoryStr: %s\n", ref.Context().RepositoryStr())
		repository := ref.Context().RepositoryStr()
		tags, err := remote.List(ref.Context(), remote.WithPageSize(10000))
		if err != nil {
			log.Infof("Couldn't list repository %s, ignoring.\n", repository)
			continue
		}

		newestVersion := ref.Identifier()
		for _, tag := range tags {
			validNewerVersion, err := isNewerVersion(newestVersion, tag, tagRegex)
			if err != nil {
				return cmd.Original, fmt.Errorf("failed to parse version %s: %v", tag, err)
			}
			if validNewerVersion {
				newestVersion = tag
			}
		}
		log.Infof("Newest version found: %s\n", newestVersion)
		newTag := ref.Context().Tag(newestVersion).Name()
		log.Infof("Newest tag found: %s\n", newTag)
		if newTag == ref.Name() {
			log.Info("No version change, not touching FROM line.")
			continue
		}
		newFrom := strings.ReplaceAll(cmd.Original, ref.Name(), newTag)
		return newFrom, nil
	}
	return cmd.Original, nil
}

func isNewerVersion(newestVersion, tag string, tagRegex string) (bool, error) {
	tagRe, err := regexp.Compile(tagRegex)
	if err != nil {
		return false, fmt.Errorf("can't parse tag regex: %v", err)
	}
	if !tagRe.Match([]byte(tag)) {
		log.Tracef("Tag %s doesn't match tag regex %s. Ignoring.\n", tag, tagRegex)
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
