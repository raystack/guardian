package usecases_test

import (
	"fmt"
	"testing"

	"github.com/odpf/guardian/repositories"
	"github.com/odpf/guardian/repositories/mocks"
	"github.com/odpf/guardian/usecases"
	"github.com/stretchr/testify/suite"
)

type UseCaseTestSuite struct {
	suite.Suite
	mockProviderRepository *mocks.Provider
	useCases               *usecases.UseCases
}

func (s *UseCaseTestSuite) SetupTest() {
	fmt.Println("setup test ===============")
	s.mockProviderRepository = new(mocks.Provider)

	allRepositories := &repositories.Repositories{
		Provider: s.mockProviderRepository,
	}
	s.useCases = usecases.New(allRepositories)
}

func TestUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}
