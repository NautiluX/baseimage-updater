package registry

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	log "github.com/sirupsen/logrus"
)

type Querier struct {
	Registry  string
	reference name.Reference
}

func NewQuerier(registry string) (*Querier, error) {
	ref, err := name.ParseReference(registry)
	if err != nil {
		return &Querier{}, fmt.Errorf("couldn't parse image %s: %v", registry, err)
	}
	r := Querier{registry, ref}
	return &r, nil
}

func (q *Querier) ListTags() (tags []string, err error) {
	log.Infof("Identifier: %s\n", q.reference.Identifier())
	log.Infof("Name: %s\n", q.reference.Name())
	log.Infof("RegistryStr: %s\n", q.reference.Context().RegistryStr())
	log.Infof("RepositoryStr: %s\n", q.reference.Context().RepositoryStr())
	repository := q.reference.Context().RepositoryStr()
	tags, err = remote.List(q.reference.Context(), remote.WithPageSize(10000))
	if err != nil {
		err = fmt.Errorf("couldn't list repository %s: %v", repository, err)
		return
	}
	return
}

func (q *Querier) GetTag() string {
	return q.reference.Identifier()
}

func (q *Querier) GetName() string {
	return q.reference.Name()
}

func (q *Querier) GetFullTag(tag string) string {
	return q.reference.Context().Tag(tag).Name()
}
