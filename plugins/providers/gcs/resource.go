package gcs

import "github.com/odpf/guardian/domain"

const (
	ResourceTypeBucket = "bucket"
	ResourceTypeObject = "object"
)

type Bucket struct {
	Name string
}

// func (b *Bucket) fromDomain(r *domain.Resource) error {
// 	if r.Type == ResourceTypeBucket {
// 		return ErrInvalidResourceType
// 	}

// 	b.Name = r.URN
// 	return nil
// }

func (b *Bucket) toDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeBucket,
		URN:  b.Name,
		Name: b.Name,
	}
}

type Object struct {
	Name string
}

// func (o *Object) fromDomain(r *domain.Resource) error {
// 	if r.Type == ResourceTypeObject {
// 		return ErrInvalidResourceType
// 	}

// 	o.Name = r.URN
// 	return nil
// }

func (o *Object) toDomain() *domain.Resource {
	return &domain.Resource{
		Type: ResourceTypeBucket,
		URN:  o.Name,
		Name: o.Name,
	}
}
