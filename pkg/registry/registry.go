package registry

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	log "github.com/sirupsen/logrus"
)

type Querier struct {
}

func NewQuerier() *Querier {
	r := Querier{}
	return &r
}

func (q *Querier) ListTags(registry string) (tags []string, err error) {
	log.Infof("Identifier: %s\n", q.reference(registry).Identifier())
	log.Infof("Name: %s\n", q.reference(registry).Name())
	log.Infof("RegistryStr: %s\n", q.reference(registry).Context().RegistryStr())
	log.Infof("RepositoryStr: %s\n", q.reference(registry).Context().RepositoryStr())
	repository := q.reference(registry).Context().RepositoryStr()
	tags, err = remote.List(q.reference(registry).Context(), remote.WithPageSize(10000))
	if err != nil {
		err = fmt.Errorf("couldn't list repository %s: %v", repository, err)
		return
	}
	return
}

func (q *Querier) GetTag(registry string) string {
	return q.reference(registry).Identifier()
}

func (q *Querier) GetName(registry string) string {
	return q.reference(registry).Name()
}

func (q *Querier) GetFullTag(registry string, tag string) string {
	return q.reference(registry).Context().Tag(tag).Name()
}

func (q *Querier) reference(registry string) name.Reference {
	ref, err := name.ParseReference(registry)
	if err != nil {
		log.Errorf("couldn't parse image %s: %v", registry, err)
	}
	return ref
}
