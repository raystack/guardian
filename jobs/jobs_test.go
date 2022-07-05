package jobs

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	. "github.com/odpf/guardian/jobs/mocks"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/identities"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type JobsTestSuite struct {
	suite.Suite
	mockAppealService   *AppealService
	mockProviderService *ProviderService
	mockPolicyService   *PolicyService
	notifier            notifiers.Client
	mockIamManager      *IamManager
	mockIamClient       *IamClient

	handler *handler
}

func (j *JobsTestSuite) SetupTest() {
	j.mockAppealService = new(AppealService)
	j.mockProviderService = new(ProviderService)
	j.mockPolicyService = new(PolicyService)
	j.notifier = new(mocks.Notifier)
	j.mockIamManager = new(IamManager)
	j.mockIamClient = new(IamClient)

	j.handler = NewHandler(log.NewLogrus(), j.mockAppealService, j.mockProviderService, j.mockPolicyService, j.notifier, j.mockIamManager)
}

func (j *JobsTestSuite) TestRun() {
	appeal1 := domain.Appeal{ID: "A1", ResourceID: "R1", PolicyID: "P1", PolicyVersion: uint(1), Status: "active", AccountType: "user", AccountID: "xyz@gojek.com"}
	appeal2 := domain.Appeal{ID: "A2", ResourceID: "R2", PolicyID: "P1", PolicyVersion: uint(1), Status: "active", AccountType: "user", AccountID: "xyz@gojek.com"}
	appeal3 := domain.Appeal{ID: "A3", ResourceID: "R1", PolicyID: "P1", PolicyVersion: uint(1), Status: "active", AccountType: "user", AccountID: "abc@gojek.com"}
	appeal4 := domain.Appeal{ID: "A4", ResourceID: "R1", PolicyID: "P1", PolicyVersion: uint(1), Status: "active", AccountType: "user", AccountID: "test@gojek.com"}
	j.mockAppealService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Appeal{&appeal1, &appeal2, &appeal3, &appeal4}, nil)

	j.mockPolicyService.On("GetOne", mock.Anything, "P1", uint(1)).Return(
		&domain.Policy{ID: "P1", Version: uint(1), IAM: &domain.IAMConfig{}}, nil)

	j.mockIamManager.On("ParseConfig", mock.Anything).Return(&identities.HTTPClientConfig{}, nil)
	j.mockIamManager.On("GetClient", mock.Anything).Return(j.mockIamClient, nil)
	j.mockIamClient.On("IsActiveUser", "xyz@gojek.com").Return(false, nil)
	j.mockIamClient.On("IsActiveUser", "abc@gojek.com").Return(true, nil)
	j.mockIamClient.On("IsActiveUser", "test@gojek.com").Return(false, nil)
	j.mockAppealService.On("Revoke", mock.Anything, "A1", domain.SystemActorName, "Automatically revoked since account is dormant").Return(&appeal1, nil)
	j.mockAppealService.On("Revoke", mock.Anything, "A2", domain.SystemActorName, "Automatically revoked since account is dormant").Return(&appeal2, nil)
	j.mockAppealService.On("Revoke", mock.Anything, "A4", domain.SystemActorName, "Automatically revoked since account is dormant").Return(&appeal4, errors.New("failed"))

	err := j.handler.DormantAccountAppealRevoke(context.Background())
	j.Nil(err)
}

func (j *JobsTestSuite) TestNoActiveAppealRun() {
	j.mockAppealService.On("Find", mock.Anything, mock.Anything).Return([]*domain.Appeal{}, nil)
	err := j.handler.DormantAccountAppealRevoke(context.Background())
	j.Nil(err)
}

func TestService(t *testing.T) {
	suite.Run(t, new(JobsTestSuite))
}
