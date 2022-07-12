package v1beta1_test

import (
	"testing"

	"github.com/odpf/guardian/api/handler/v1beta1"
	"github.com/odpf/guardian/api/handler/v1beta1/mocks"
	"github.com/stretchr/testify/suite"
)

type GrpcHandlersSuite struct {
	suite.Suite

	resourceService *mocks.ResourceService
	providerService *mocks.ProviderService
	policyService   *mocks.PolicyService
	appealService   *mocks.AppealService
	approvalService *mocks.ApprovalService
	grpcServer      *v1beta1.GRPCServer

	authenticatedUserHeaderKey string
}

func TestGrpcHandler(t *testing.T) {
	suite.Run(t, new(GrpcHandlersSuite))
}

func (s *GrpcHandlersSuite) setup() {
	s.resourceService = new(mocks.ResourceService)
	s.providerService = new(mocks.ProviderService)
	s.policyService = new(mocks.PolicyService)
	s.appealService = new(mocks.AppealService)
	s.approvalService = new(mocks.ApprovalService)
	s.authenticatedUserHeaderKey = "test-header-key"
	s.grpcServer = v1beta1.NewGRPCServer(
		s.resourceService,
		s.providerService,
		s.policyService,
		s.appealService,
		s.approvalService,
		v1beta1.NewAdapter(),
		s.authenticatedUserHeaderKey,
	)
}
