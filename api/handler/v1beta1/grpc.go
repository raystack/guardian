package v1beta1

import (
	"context"
	"errors"

	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/core/grant"

	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/domain"
	"google.golang.org/grpc/metadata"
)

type ProtoAdapter interface {
	FromProviderProto(*guardianv1beta1.Provider) (*domain.Provider, error)
	FromProviderConfigProto(*guardianv1beta1.ProviderConfig) *domain.ProviderConfig
	ToProviderProto(*domain.Provider) (*guardianv1beta1.Provider, error)
	ToProviderConfigProto(*domain.ProviderConfig) (*guardianv1beta1.ProviderConfig, error)
	ToProviderTypeProto(domain.ProviderType) *guardianv1beta1.ProviderType
	ToRole(*domain.Role) (*guardianv1beta1.Role, error)

	FromPolicyProto(*guardianv1beta1.Policy) *domain.Policy
	ToPolicyProto(*domain.Policy) (*guardianv1beta1.Policy, error)

	ToPolicyAppealConfigProto(policy *domain.Policy) *guardianv1beta1.PolicyAppealConfig

	FromResourceProto(*guardianv1beta1.Resource) *domain.Resource
	ToResourceProto(*domain.Resource) (*guardianv1beta1.Resource, error)

	ToAppealProto(*domain.Appeal) (*guardianv1beta1.Appeal, error)
	FromCreateAppealProto(*guardianv1beta1.CreateAppealRequest, string) ([]*domain.Appeal, error)
	ToApprovalProto(*domain.Approval) (*guardianv1beta1.Approval, error)

	ToGrantProto(*domain.Grant) (*guardianv1beta1.Grant, error)
	FromGrantProto(*guardianv1beta1.Grant) *domain.Grant
}

//go:generate mockery --name=resourceService --exported --with-expecter
type resourceService interface {
	Find(context.Context, domain.ListResourcesFilter) ([]*domain.Resource, error)
	GetOne(context.Context, string) (*domain.Resource, error)
	BulkUpsert(context.Context, []*domain.Resource) error
	Update(context.Context, *domain.Resource) error
	Get(context.Context, *domain.ResourceIdentifier) (*domain.Resource, error)
	Delete(context.Context, string) error
	BatchDelete(context.Context, []string) error
}

//go:generate mockery --name=providerService --exported --with-expecter
type providerService interface {
	Create(context.Context, *domain.Provider) error
	Find(context.Context) ([]*domain.Provider, error)
	GetByID(context.Context, string) (*domain.Provider, error)
	GetTypes(context.Context) ([]domain.ProviderType, error)
	GetOne(ctx context.Context, pType, urn string) (*domain.Provider, error)
	Update(context.Context, *domain.Provider) error
	FetchResources(context.Context) error
	GetRoles(ctx context.Context, id, resourceType string) ([]*domain.Role, error)
	ValidateAppeal(context.Context, *domain.Appeal, *domain.Provider, *domain.Policy) error
	GrantAccess(context.Context, domain.Grant) error
	RevokeAccess(context.Context, domain.Grant) error
	Delete(context.Context, string) error
}

//go:generate mockery --name=policyService --exported --with-expecter
type policyService interface {
	Create(context.Context, *domain.Policy) error
	Find(context.Context) ([]*domain.Policy, error)
	GetOne(ctx context.Context, id string, version uint) (*domain.Policy, error)
	Update(context.Context, *domain.Policy) error
}

//go:generate mockery --name=appealService --exported --with-expecter
type appealService interface {
	GetByID(context.Context, string) (*domain.Appeal, error)
	Find(context.Context, *domain.ListAppealsFilter) ([]*domain.Appeal, error)
	Create(context.Context, []*domain.Appeal, ...appeal.CreateAppealOption) error
	Cancel(context.Context, string) (*domain.Appeal, error)
	AddApprover(ctx context.Context, appealID, approvalID, email string) (*domain.Appeal, error)
	DeleteApprover(ctx context.Context, appealID, approvalID, email string) (*domain.Appeal, error)
	UpdateApproval(ctx context.Context, approvalAction domain.ApprovalAction) (*domain.Appeal, error)
}

//go:generate mockery --name=approvalService --exported --with-expecter
type approvalService interface {
	ListApprovals(context.Context, *domain.ListApprovalsFilter) ([]*domain.Approval, error)
	BulkInsert(context.Context, []*domain.Approval) error
}

//go:generate mockery --name=grantService --exported --with-expecter
type grantService interface {
	List(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)
	GetByID(context.Context, string) (*domain.Grant, error)
	Revoke(ctx context.Context, id, actor, reason string, opts ...grant.Option) (*domain.Grant, error)
	BulkRevoke(ctx context.Context, filter domain.RevokeGrantsFilter, actor, reason string) ([]*domain.Grant, error)
	ImportFromProvider(ctx context.Context, criteria grant.ImportFromProviderCriteria) ([]*domain.Grant, error)
}

type GRPCServer struct {
	resourceService resourceService
	providerService providerService
	policyService   policyService
	appealService   appealService
	approvalService approvalService
	grantService    grantService
	adapter         ProtoAdapter

	authenticatedUserHeaderKey string

	guardianv1beta1.UnimplementedGuardianServiceServer
}

func NewGRPCServer(
	resourceService resourceService,
	providerService providerService,
	policyService policyService,
	appealService appealService,
	approvalService approvalService,
	grantService grantService,
	adapter ProtoAdapter,
	authenticatedUserHeaderKey string,
) *GRPCServer {
	return &GRPCServer{
		resourceService:            resourceService,
		providerService:            providerService,
		policyService:              policyService,
		appealService:              appealService,
		approvalService:            approvalService,
		grantService:               grantService,
		adapter:                    adapter,
		authenticatedUserHeaderKey: authenticatedUserHeaderKey,
	}
}

func (s *GRPCServer) getUser(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("unable to retrieve metadata from context")
	}

	users := md.Get(s.authenticatedUserHeaderKey)
	if len(users) == 0 {
		return "", errors.New("user email not found")
	}

	currentUser := users[0]
	return currentUser, nil
}
