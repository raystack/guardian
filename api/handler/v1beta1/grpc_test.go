package v1beta1_test

import (
	"testing"

	"github.com/raystack/guardian/api/handler/v1beta1"
	"github.com/raystack/guardian/api/handler/v1beta1/mocks"
	"github.com/stretchr/testify/suite"
)

type authEmailTestContextKey struct{}

type GrpcHandlersSuite struct {
	suite.Suite

	resourceService *mocks.ResourceService
	activityService *mocks.ActivityService
	providerService *mocks.ProviderService
	policyService   *mocks.PolicyService
	appealService   *mocks.AppealService
	approvalService *mocks.ApprovalService
	grantService    *mocks.GrantService
	grpcServer      *v1beta1.GRPCServer
}

func TestGrpcHandler(t *testing.T) {
	suite.Run(t, new(GrpcHandlersSuite))
}

func (s *GrpcHandlersSuite) setup() {
	s.resourceService = new(mocks.ResourceService)
	s.activityService = new(mocks.ActivityService)
	s.providerService = new(mocks.ProviderService)
	s.policyService = new(mocks.PolicyService)
	s.appealService = new(mocks.AppealService)
	s.approvalService = new(mocks.ApprovalService)
	s.grantService = new(mocks.GrantService)
	s.grpcServer = v1beta1.NewGRPCServer(
		s.resourceService,
		s.activityService,
		s.providerService,
		s.policyService,
		s.appealService,
		s.approvalService,
		s.grantService,
		v1beta1.NewAdapter(),
		authEmailTestContextKey{},
	)
}
