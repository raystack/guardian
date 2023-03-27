package resource_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/goto/guardian/core/resource"
	"github.com/goto/guardian/core/resource/mocks"
	"github.com/goto/guardian/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockRepository  *mocks.Repository
	mockAuditLogger *mocks.AuditLogger
	service         *resource.Service

	authenticatedUserEmail string
}

func (s *ServiceTestSuite) SetupTest() {
	s.mockRepository = new(mocks.Repository)
	s.mockAuditLogger = new(mocks.AuditLogger)
	s.service = resource.NewService(resource.ServiceDeps{
		Repository:  s.mockRepository,
		AuditLogger: s.mockAuditLogger,
	})
	s.authenticatedUserEmail = "user@example.com"
}

func (s *ServiceTestSuite) TestFind() {
	s.Run("should return nil and error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(nil, expectedError).Once()

		actualResult, actualError := s.service.Find(context.Background(), domain.ListResourcesFilter{})

		s.Nil(actualResult)
		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should return list of records on success", func() {
		expectedFilters := domain.ListResourcesFilter{}
		expectedResult := []*domain.Resource{}
		s.mockRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), expectedFilters).Return(expectedResult, nil).Once()

		actualResult, actualError := s.service.Find(context.Background(), expectedFilters)

		s.Equal(expectedResult, actualResult)
		s.Nil(actualError)
		s.mockRepository.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) TestBulkUpsert() {
	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.EXPECT().BulkUpsert(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).Return(expectedError).Once()
		s.mockAuditLogger.EXPECT().Log(mock.Anything, resource.AuditKeyResoruceBulkUpsert, mock.Anything).Return(nil)

		actualError := s.service.BulkUpsert(context.Background(), []*domain.Resource{})

		s.EqualError(actualError, expectedError.Error())
	})
}

func (s *ServiceTestSuite) TestUpdate() {
	s.Run("should return error if got error getting existing record", func() {
		testCases := []struct {
			expectedExistingResource *domain.Resource
			expectedRepositoryError  error
			expectedError            error
		}{
			{
				expectedExistingResource: nil,
				expectedRepositoryError:  resource.ErrRecordNotFound,
				expectedError:            resource.ErrRecordNotFound,
			},
			{
				expectedExistingResource: nil,
				expectedRepositoryError:  errors.New("repository error"),
				expectedError:            errors.New("repository error"),
			},
		}

		for _, tc := range testCases {
			expectedResource := &domain.Resource{
				ID: "1",
			}
			expectedError := tc.expectedError
			s.mockRepository.EXPECT().
				GetOne(mock.AnythingOfType("*context.emptyCtx"), expectedResource.ID).
				Return(tc.expectedExistingResource, tc.expectedRepositoryError).Once()

			actualError := s.service.Update(context.Background(), expectedResource)

			s.EqualError(actualError, expectedError.Error())
		}
	})

	s.Run("should return error if got error from repository", func() {
		expectedError := errors.New("error from repository")
		s.mockRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(&domain.Resource{}, nil).Once()
		s.mockRepository.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
			Return(expectedError).Once()

		actualError := s.service.Update(context.Background(), &domain.Resource{})

		s.EqualError(actualError, expectedError.Error())
	})

	s.Run("should only allows details and labels to be edited", func() {
		testCases := []struct {
			name                  string
			resourceUpdatePayload *domain.Resource
			existingResource      *domain.Resource
			expectedUpdatedValues *domain.Resource
		}{
			{
				name: "empty labels in existing resource",
				resourceUpdatePayload: &domain.Resource{
					ID: "1",
					Labels: map[string]string{
						"key": "value",
					},
				},
				existingResource: &domain.Resource{
					ID: "1",
				},
				expectedUpdatedValues: &domain.Resource{
					ID: "1",
					Labels: map[string]string{
						"key": "value",
					},
				},
			},
			{
				name: "empty details in existing resource",
				resourceUpdatePayload: &domain.Resource{
					ID: "2",
					Details: map[string]interface{}{
						"key": "value",
					},
				},
				existingResource: &domain.Resource{
					ID: "2",
				},
				expectedUpdatedValues: &domain.Resource{
					ID: "2",
					Details: map[string]interface{}{
						"key": "value",
					},
				},
			},
			{
				name: "trying to update resource type",
				resourceUpdatePayload: &domain.Resource{
					ID:   "2",
					Type: "test",
				},
				existingResource: &domain.Resource{
					ID: "2",
				},
				expectedUpdatedValues: &domain.Resource{
					ID: "2",
				},
			},
			{
				name: "should exclude __metadata from update payload",
				resourceUpdatePayload: &domain.Resource{
					ID: "2",
					Details: map[string]interface{}{
						"owner": "new-owner@example.com",
						resource.ReservedDetailsKeyMetadata: map[string]string{
							"new-key": "new-value",
						},
					},
				},
				existingResource: &domain.Resource{
					ID: "2",
					Details: map[string]interface{}{
						"owner": "user@example.com",
						"foo":   "bar",
						resource.ReservedDetailsKeyMetadata: map[string]string{
							"key": "value",
						},
					},
				},
				expectedUpdatedValues: &domain.Resource{
					ID: "2",
					Details: map[string]interface{}{
						"owner": "new-owner@example.com",
						"foo":   "bar",
						resource.ReservedDetailsKeyMetadata: map[string]string{
							"key": "value",
						},
					},
				},
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				s.mockRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), tc.resourceUpdatePayload.ID).Return(tc.existingResource, nil).Once()
				s.mockRepository.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*domain.Resource")).
					Run(func(_a0 context.Context, updateResourcePayload *domain.Resource) {
						s.Empty(cmp.Diff(tc.expectedUpdatedValues, updateResourcePayload, cmpopts.IgnoreFields(domain.Resource{}, "UpdatedAt", "CreatedAt")))
					}).Return(nil).Once()
				s.mockAuditLogger.EXPECT().Log(mock.Anything, resource.AuditKeyResourceUpdate, mock.Anything).Return(nil)

				actualError := s.service.Update(context.Background(), tc.resourceUpdatePayload)

				s.Nil(actualError)
				s.mockRepository.AssertExpectations(s.T())
			})
		}
	})
}

func (s *ServiceTestSuite) TestGet() {
	s.Run("success scenarios", func() {
		s.Run("should return resource details when using resource id", func() {
			expectedResource := &domain.Resource{
				ID: "1",
			}
			s.mockRepository.EXPECT().GetOne(mock.AnythingOfType("*context.emptyCtx"), expectedResource.ID).
				Return(expectedResource, nil).Once()

			actualResource, actualError := s.service.Get(context.Background(), &domain.ResourceIdentifier{ID: expectedResource.ID})

			s.Nil(actualError)
			s.Equal(expectedResource, actualResource)
		})

		s.Run("should return resource details when using resource urn", func() {
			expectedResource := &domain.Resource{
				ID:           "1",
				ProviderType: "test-provider",
				ProviderURN:  "test-provider-urn",
				Type:         "test-type",
				URN:          "test-urn",
			}
			s.mockRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), domain.ListResourcesFilter{
				ProviderType: "test-provider",
				ProviderURN:  "test-provider-urn",
				ResourceType: "test-type",
				ResourceURN:  "test-urn",
			}).
				Return([]*domain.Resource{expectedResource}, nil).Once()

			actualResource, actualError := s.service.Get(context.Background(), &domain.ResourceIdentifier{
				ProviderType: expectedResource.ProviderType,
				ProviderURN:  expectedResource.ProviderURN,
				Type:         expectedResource.Type,
				URN:          expectedResource.URN,
			})

			s.Nil(actualError)
			s.Equal(expectedResource, actualResource)
		})
	})

	s.Run("should return not found if resource not found", func() {
		expectedError := resource.ErrRecordNotFound
		s.mockRepository.EXPECT().Find(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("domain.ListResourcesFilter")).
			Return([]*domain.Resource{}, nil).Once()

		actualResource, actualError := s.service.Get(context.Background(), &domain.ResourceIdentifier{
			ProviderType: "test-provider",
			ProviderURN:  "test-provider-urn",
			Type:         "test-type",
			URN:          "test-urn",
		})

		s.ErrorIs(actualError, expectedError)
		s.Nil(actualResource)
	})
}

func (s *ServiceTestSuite) TestDelete() {
	s.Run("should delete resource", func() {
		expectedResourceID := "test-resource-id"

		s.mockRepository.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), expectedResourceID).
			Return(nil).Once()
		s.mockAuditLogger.EXPECT().Log(mock.Anything, resource.AuditKeyResourceDelete, mock.Anything).Return(nil)

		actualError := s.service.Delete(context.Background(), expectedResourceID)

		s.Nil(actualError)
	})

	s.Run("should return error if repository returns an error", func() {
		expectedResourceID := "test-resource-id"
		expectedError := errors.New("test-error")

		s.mockRepository.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), expectedResourceID).
			Return(expectedError).Once()

		actualError := s.service.Delete(context.Background(), expectedResourceID)

		s.ErrorIs(actualError, expectedError)
	})
}

func (s *ServiceTestSuite) TestBatchDelete() {
	s.Run("should delete resources", func() {
		expectedResourceIDs := []string{"test-resource-id"}

		s.mockRepository.EXPECT().BatchDelete(mock.AnythingOfType("*context.emptyCtx"), expectedResourceIDs).
			Return(nil).Once()
		s.mockAuditLogger.EXPECT().Log(mock.Anything, resource.AuditKeyResourceBatchDelete, mock.Anything).Return(nil)

		actualError := s.service.BatchDelete(context.Background(), expectedResourceIDs)

		s.Nil(actualError)
	})

	s.Run("should return error if repository returns an error", func() {
		expectedResourceIDs := []string{"test-resource-id"}
		expectedError := errors.New("test-error")

		s.mockRepository.EXPECT().BatchDelete(mock.AnythingOfType("*context.emptyCtx"), expectedResourceIDs).
			Return(expectedError).Once()

		actualError := s.service.BatchDelete(context.Background(), expectedResourceIDs)

		s.ErrorIs(actualError, expectedError)
	})
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
