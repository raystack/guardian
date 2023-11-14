package v1beta1

import (
	"context"
	"strings"

	"github.com/raystack/guardian/core/appeal"
	"github.com/raystack/guardian/core/grant"

	guardianv1beta1 "github.com/raystack/guardian/api/proto/raystack/guardian/v1beta1"
	"github.com/raystack/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProtoAdapter interface {
	FromProviderProto(*guardianv1beta1.Provider) (*domain.Provider, error)
	FromProviderConfigProto(*guardianv1beta1.ProviderConfig) *domain.ProviderConfig
	ToProviderProto(*domain.Provider) (*guardianv1beta1.Provider, error)
	ToProviderConfigProto(*domain.ProviderConfig) (*guardianv1beta1.ProviderConfig, error)
	ToProviderTypeProto(domain.ProviderType) *guardianv1beta1.ProviderType
	ToRole(*domain.Role) (*guardianv1beta1.Role, error)

	FromPolicyProto(*guardianv1beta1.Policy, string) *domain.Policy
	ToPolicyProto(*domain.Policy) (*guardianv1beta1.Policy, error)

	ToPolicyAppealConfigProto(policy *domain.Policy) *guardianv1beta1.PolicyAppealConfig

	FromResourceProto(*guardianv1beta1.Resource) *domain.Resource
	ToResourceProto(*domain.Resource) (*guardianv1beta1.Resource, error)

	ToAppealProto(*domain.Appeal) (*guardianv1beta1.Appeal, error)
	FromCreateAppealProto(*guardianv1beta1.CreateAppealRequest, string) ([]*domain.Appeal, error)
	ToApprovalProto(*domain.Approval) (*guardianv1beta1.Approval, error)

	ToGrantProto(*domain.Grant) (*guardianv1beta1.Grant, error)
	FromGrantProto(*guardianv1beta1.Grant) *domain.Grant

	ToActivityProto(*domain.Activity) (*guardianv1beta1.ProviderActivity, error)
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

//go:generate mockery --name=activityService --exported --with-expecter
type activityService interface {
	GetOne(context.Context, string) (*domain.Activity, error)
	Find(context.Context, domain.ListProviderActivitiesFilter) ([]*domain.Activity, error)
	Import(context.Context, domain.ListActivitiesFilter) ([]*domain.Activity, error)
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
	GetAppealsTotalCount(context.Context, *domain.ListAppealsFilter) (int64, error)
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
	GetApprovalsTotalCount(context.Context, *domain.ListApprovalsFilter) (int64, error)
	BulkInsert(context.Context, []*domain.Approval) error
}

//go:generate mockery --name=grantService --exported --with-expecter
type grantService interface {
	GetGrantsTotalCount(context.Context, domain.ListGrantsFilter) (int64, error)
	List(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)
	GetByID(context.Context, string) (*domain.Grant, error)
	Update(context.Context, *domain.Grant) error
	Revoke(ctx context.Context, id, actor, reason string, opts ...grant.Option) (*domain.Grant, error)
	BulkRevoke(ctx context.Context, filter domain.RevokeGrantsFilter, actor, reason string) ([]*domain.Grant, error)
	ImportFromProvider(ctx context.Context, criteria grant.ImportFromProviderCriteria) ([]*domain.Grant, error)
}

//go:generate mockery --name=namespaceService --exported --with-expecter
type namespaceService interface {
	Get(ctx context.Context, id string) (*domain.Namespace, error)
	Create(ctx context.Context, namespace *domain.Namespace) error
	Update(ctx context.Context, namespace *domain.Namespace) error
	List(ctx context.Context, filter domain.NamespaceFilter) ([]*domain.Namespace, error)
}

type GRPCServer struct {
	resourceService  resourceService
	activityService  activityService
	providerService  providerService
	policyService    policyService
	appealService    appealService
	approvalService  approvalService
	grantService     grantService
	namespaceService namespaceService
	adapter          ProtoAdapter

	authenticatedUserContextKey interface{}

	guardianv1beta1.UnimplementedGuardianServiceServer
}

func NewGRPCServer(
	resourceService resourceService,
	activityService activityService,
	providerService providerService,
	policyService policyService,
	appealService appealService,
	approvalService approvalService,
	grantService grantService,
	namespaceService namespaceService,
	adapter ProtoAdapter,
	authenticatedUserContextKey interface{},
) *GRPCServer {
	return &GRPCServer{
		resourceService:             resourceService,
		activityService:             activityService,
		providerService:             providerService,
		policyService:               policyService,
		appealService:               appealService,
		approvalService:             approvalService,
		grantService:                grantService,
		namespaceService:            namespaceService,
		adapter:                     adapter,
		authenticatedUserContextKey: authenticatedUserContextKey,
	}
}

func (s *GRPCServer) getUser(ctx context.Context) (string, error) {
	authenticatedEmail, ok := ctx.Value(s.authenticatedUserContextKey).(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "unable to get authenticated user from context")
	}

	if strings.TrimSpace(authenticatedEmail) == "" {
		return "", status.Error(codes.Unauthenticated, "unable to get authenticated user from context")
	}

	return authenticatedEmail, nil
}
