package gcs

import "github.com/odpf/guardian/domain"

const (
	ResourceTypeBucket = "bucket"
)

type Bucket struct {
	Name string
}

func (b *Bucket) fromDomain(r *domain.Resource) error {
	if r.Type != ResourceTypeBucket {
		return ErrInvalidResourceType
	}

	b.Name = r.URN
	return nil
}

func (b *Bucket) toDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeBucket,
		URN:  b.Name,
		Name: b.Name,
	}
}
