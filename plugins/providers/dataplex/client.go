package dataplex

import (
	"context"
	"fmt"
	"strings"

	datacatalog "cloud.google.com/go/datacatalog/apiv1"
	"github.com/raystack/guardian/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/cloud/datacatalog/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
)

type policyTagClient struct {
	policyManager    *datacatalog.PolicyTagManagerClient
	projectId        string
	taxonomyLocation string
}

func newPolicyTagClient(projectID, location string, credentialsJSON []byte) (*policyTagClient, error) {
	ctx := context.Background()
	policyManager, err := datacatalog.NewPolicyTagManagerClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, err
	}
	return &policyTagClient{
		policyManager:    policyManager,
		projectId:        projectID,
		taxonomyLocation: location,
	}, nil
}

func (p *policyTagClient) GetPolicies(ctx context.Context) ([]*Policy, error) {
	taxonomies := p.policyManager.ListTaxonomies(ctx, &pb.ListTaxonomiesRequest{
		Parent: p.toParent(p.projectId, p.taxonomyLocation),
	})
	taxonomies.PageInfo().MaxSize = PageSize

	policyTags := make([]*Policy, 0)
	for {
		taxonomy, err := taxonomies.Next()
		if err == iterator.Done {
			break
		}
		tags := p.policyManager.ListPolicyTags(ctx, &pb.ListPolicyTagsRequest{
			Parent: taxonomy.Name,
		})
		tags.PageInfo().MaxSize = PageSize

		for {
			tag, err := tags.Next()
			if err == iterator.Done {
				break
			}
			policyTags = append(policyTags, &Policy{
				Name:                tag.Name,
				DisplayName:         tag.DisplayName,
				Description:         tag.Description,
				TaxonomyDisplayName: taxonomy.DisplayName,
			})
		}
	}
	return policyTags, nil
}

func (p *policyTagClient) GrantPolicyAccess(ctx context.Context, tag *Policy, user, role string) error {
	policy, err := p.policyManager.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: tag.Name,
	})
	if err != nil {
		return err
	}
	usersWithGivenRole := make([]string, 0)
	for _, bind := range policy.Bindings {
		if role == bind.Role {
			usersWithGivenRole = append(bind.Members, user)
			for _, member := range bind.Members {
				if member == user {
					return ErrPermissionAlreadyExists
				}
			}
		}
	}
	if len(usersWithGivenRole) == 0 {
		usersWithGivenRole = append(usersWithGivenRole, user)
	}

	_, err = p.policyManager.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: tag.Name,
		Policy: &iampb.Policy{Bindings: []*iampb.Binding{{
			Role:    role,
			Members: usersWithGivenRole,
		}}},
	})

	return err
}

func (p *policyTagClient) RevokePolicyAccess(ctx context.Context, tag *Policy, user, role string) error {
	policy, err := p.policyManager.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: tag.Name,
	})
	if err != nil {
		return err
	}

	usersWithGivenRole := make([]string, 0)
	for _, bind := range policy.Bindings {
		if role == bind.Role {
			for _, member := range bind.Members {
				if member == user {
					continue
				}
				usersWithGivenRole = append(usersWithGivenRole, member)
			}

			if len(usersWithGivenRole) == len(bind.Members) {
				return ErrPermissionNotFound
			}
		}
	}

	_, err = p.policyManager.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: tag.Name,
		Policy: &iampb.Policy{Bindings: []*iampb.Binding{{
			Role:    role,
			Members: usersWithGivenRole,
		}}},
	})
	return err
}

func (p *policyTagClient) ListAccess(ctx context.Context, resources []*domain.Resource) (domain.MapResourceAccess, error) {
	access := make(domain.MapResourceAccess)

	for _, r := range resources {
		var accessEntries []domain.AccessEntry

		switch r.Type {
		case ResourceTypeTag:
			policy := new(Policy)
			policy.FromDomain(r)

			say, err := p.policyManager.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
				Resource: policy.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("getting policy-tag access entries of %q, %w", r.URN, err)
			}

			for _, binding := range say.Bindings {
				for _, member := range binding.Members {
					accountType := ""
					if strings.HasPrefix(member, AccountTypeUser) {
						accountType = AccountTypeUser
					} else if strings.HasPrefix(member, AccountTypeServiceAccount) {
						accountType = AccountTypeServiceAccount
					}
					if len(accountType) == 0 {
						accessEntries = append(accessEntries, domain.AccessEntry{
							AccountID:   member,
							AccountType: accountType,
							Permission:  binding.GetRole(),
						})
					}
				}
			}
		}
		if accessEntries != nil {
			access[r.URN] = accessEntries
		}
	}
	return access, nil
}

func (p *policyTagClient) toParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func containsString(arr []string, v string) bool {
	for _, item := range arr {
		if item == v {
			return true
		}
	}
	return false
}
