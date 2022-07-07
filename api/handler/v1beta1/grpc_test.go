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
	grpcServer      *v1beta1.GRPCServer
}

func TestGrpcHandler(t *testing.T) {
	suite.Run(t, new(GrpcHandlersSuite))
}

func (s *GrpcHandlersSuite) setup() {
	s.resourceService = new(mocks.ResourceService)
	s.grpcServer = v1beta1.NewGRPCServer(
		s.resourceService,
		nil,
		nil,
		nil,
		nil,
		v1beta1.NewAdapter(),
		"",
	)
}
