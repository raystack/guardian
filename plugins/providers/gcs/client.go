package gcs

import (
	"context"
	"fmt"

	"cloud.google.com/go/iam"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

//go:generate mockery --name=GCSClient --exported --with-expecter
type GCSClient interface {
	GetBuckets(ctx context.Context, projectID string) ([]*Bucket, error)
	GrantBucketAccess(ctx context.Context, b Bucket, identity string, roleName iam.RoleName) error
	RevokeBucketAccess(ctx context.Context, b Bucket, identity string, roleName iam.RoleName) error
}

type gcsClient struct {
	client    *storage.Client
	projectID string
}

func newGCSClient(projectID string, credentialsJSON []byte) (*gcsClient, error) {
	client, err := storage.NewClient(context.TODO(), option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, err
	}

	return &gcsClient{
		client:    client,
		projectID: projectID,
	}, nil
}

//GetBuckets returns all buckets within a given project
func (c *gcsClient) GetBuckets(ctx context.Context, projectID string) ([]*Bucket, error) {
	var result []*Bucket
	it := c.client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		result = append(result, &Bucket{
			Name: battrs.Name,
		})
	}
	return result, nil
}

func (c *gcsClient) GrantBucketAccess(ctx context.Context, b Bucket, identity string, roleName iam.RoleName) error {
	bucketName := b.Name
	bucket := c.client.Bucket(bucketName)
	policy, err := bucket.IAM().Policy(ctx)
	if err != nil {
		return fmt.Errorf("Bucket(%q).IAM().Policy: %w", bucketName, err)
	}

	policy.Add(identity, roleName)
	if err := bucket.IAM().SetPolicy(ctx, policy); err != nil {
		return fmt.Errorf("Bucket(%q).IAM().SetPolicy: %w", bucketName, err)
	}

	return nil
}

func (c *gcsClient) RevokeBucketAccess(ctx context.Context, b Bucket, identity string, roleName iam.RoleName) error {
	bucketName := b.Name
	bucket := c.client.Bucket(bucketName)
	policy, err := bucket.IAM().Policy(ctx)
	if err != nil {
		return fmt.Errorf("Bucket(%q).IAM().Policy: %w", bucketName, err)
	}

	policy.Remove(identity, roleName)
	if err := bucket.IAM().SetPolicy(ctx, policy); err != nil {
		return fmt.Errorf("Bucket(%q).IAM().SetPolicy: %w", bucketName, err)
	}

	return nil
}
