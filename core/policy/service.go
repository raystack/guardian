//go:generate mockery --name=repository --exported
//go:generate mockery --name=providerService --exported
//go:generate mockery --name=resourceService --exported
//go:generate mockery --name=auditLogger --exported

package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/internal/evaluator"
	"github.com/odpf/guardian/store"
	"github.com/odpf/salt/log"
)

const (
	AuditKeyPolicyCreate = "policy.create"
	AuditKeyPolicyUpdate = "policy.update"
)

type repository interface {
	store.PolicyRepository
}

type providerService interface {
	GetOne(ctx context.Context, pType, urn string) (*domain.Provider, error)
	ValidateAppeal(context.Context, *domain.Appeal, *domain.Provider) error
}

type resourceService interface {
	Get(context.Context, *domain.ResourceIdentifier) (*domain.Resource, error)
}

type auditLogger interface {
	Log(ctx context.Context, action string, data interface{}) error
}

// Service handling the business logics
type Service struct {
	repository      store.PolicyRepository
	resourceService resourceService
	providerService providerService
	iam             domain.IAMManager

	validator   *validator.Validate
	logger      log.Logger
	auditLogger auditLogger
}

type ServiceOptions struct {
	Repository      repository
	ResourceService resourceService
	ProviderService providerService
	IAMManager      domain.IAMManager

	Validator   *validator.Validate
	Logger      log.Logger
	AuditLogger auditLogger
}

// NewService returns service struct
func NewService(opts ServiceOptions) *Service {
	return &Service{
		opts.Repository,
		opts.ResourceService,
		opts.ProviderService,
		opts.IAMManager,

		opts.Validator,
		opts.Logger,
		opts.AuditLogger,
	}
}

// Create record
func (s *Service) Create(ctx context.Context, p *domain.Policy) error {
	p.Version = 1

	var sensitiveConfig domain.SensitiveConfig
	if p.HasIAMConfig() {
		iamClientConfig, err := s.iam.ParseConfig(p.IAM)
		if err != nil {
			return fmt.Errorf("parsing iam config: %w", err)
		}
		sensitiveConfig = iamClientConfig
		p.IAM.Config = sensitiveConfig
	}

	if err := s.validatePolicy(ctx, p); err != nil {
		return fmt.Errorf("policy validation: %w", err)
	}

	if p.HasIAMConfig() {
		if err := sensitiveConfig.Encrypt(); err != nil {
			return fmt.Errorf("encrypting iam config: %w", err)
		}
		p.IAM.Config = sensitiveConfig
	}
	if err := s.repository.Create(p); err != nil {
		return err
	}

	if p.HasIAMConfig() {
		if err := s.decryptAndDeserializeIAMConfig(p.IAM); err != nil {
			return err
		}
	}

	if err := s.auditLogger.Log(ctx, AuditKeyPolicyCreate, p); err != nil {
		s.logger.Error(fmt.Sprintf("failed to record audit log: %s", err))
	}

	return nil
}

// Find records
func (s *Service) Find(ctx context.Context) ([]*domain.Policy, error) {
	policies, err := s.repository.Find()
	if err != nil {
		return nil, err
	}

	for _, p := range policies {
		if p.HasIAMConfig() {
			if err := s.decryptAndDeserializeIAMConfig(p.IAM); err != nil {
				return nil, err
			}
		}
	}
	return policies, nil
}

// GetOne record
func (s *Service) GetOne(ctx context.Context, id string, version uint) (*domain.Policy, error) {
	p, err := s.repository.GetOne(id, version)
	if err != nil {
		return nil, err
	}

	if p.HasIAMConfig() {
		if err := s.decryptAndDeserializeIAMConfig(p.IAM); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// Update a record
func (s *Service) Update(ctx context.Context, p *domain.Policy) error {
	if p.ID == "" {
		return ErrEmptyIDParam
	}

	var sensitiveConfig domain.SensitiveConfig
	if p.HasIAMConfig() {
		iamClientConfig, err := s.iam.ParseConfig(p.IAM)
		if err != nil {
			return fmt.Errorf("parsing iam config: %w", err)
		}
		sensitiveConfig = iamClientConfig
		p.IAM.Config = sensitiveConfig
	}

	if err := s.validatePolicy(ctx, p, "Version"); err != nil {
		return fmt.Errorf("policy validation: %w", err)
	}

	latestPolicy, err := s.GetOne(ctx, p.ID, p.Version)
	if err != nil {
		return err
	}

	p.Version = latestPolicy.Version + 1

	if p.HasIAMConfig() {
		if err := sensitiveConfig.Encrypt(); err != nil {
			return fmt.Errorf("encrypting iam config: %w", err)
		}
		p.IAM.Config = sensitiveConfig
	}

	if err := s.repository.Create(p); err != nil {
		return err
	}

	if err := s.auditLogger.Log(ctx, AuditKeyPolicyUpdate, p); err != nil {
		s.logger.Error(fmt.Sprintf("failed to record audit log: %s", err))
	}

	if p.HasIAMConfig() {
		if err := s.decryptAndDeserializeIAMConfig(p.IAM); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) decryptAndDeserializeIAMConfig(c *domain.IAMConfig) error {
	iamClientConfig, err := s.iam.ParseConfig(c)
	if err != nil {
		return fmt.Errorf("parsing iam config: %w", err)
	}
	if err := iamClientConfig.Decrypt(); err != nil {
		return fmt.Errorf("decrypting iam config: %w", err)
	}
	iamClientConfigMap, err := structToMap(iamClientConfig)
	if err != nil {
		return fmt.Errorf("deserializing iam config: %w", err)
	}

	c.Config = iamClientConfigMap
	return nil
}

func (s *Service) validatePolicy(ctx context.Context, p *domain.Policy, excludedFields ...string) error {
	if containsWhitespaces(p.ID) {
		return ErrIDContainsWhitespaces
	}

	if err := s.validator.StructExcept(p, excludedFields...); err != nil {
		return err
	}

	if err := s.validateSteps(p.Steps); err != nil {
		return err
	}

	if err := s.validateRequirements(ctx, p.Requirements); err != nil {
		return fmt.Errorf("invalid requirements: %w", err)
	}

	if p.HasIAMConfig() {
		if config, ok := p.IAM.Config.(domain.SensitiveConfig); ok {
			if err := config.Validate(); err != nil {
				return fmt.Errorf("invalid iam config: %w", err)
			}
		} else {
			config, err := s.iam.ParseConfig(p.IAM)
			if err != nil {
				return fmt.Errorf("parsing iam config: %w", err)
			}

			if err := config.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) validateRequirements(ctx context.Context, requirements []*domain.Requirement) error {
	for i, r := range requirements {
		for j, aa := range r.Appeals {
			resource, err := s.resourceService.Get(ctx, aa.Resource)
			if err != nil {
				return fmt.Errorf("requirement[%v].appeals[%v].resource: %w", i, j, err)
			}
			provider, err := s.providerService.GetOne(ctx, resource.ProviderType, resource.ProviderURN)
			if err != nil {
				return fmt.Errorf("requirement[%v].appeals[%v].resource: retrieving provider: %w", i, j, err)
			}

			appeal := &domain.Appeal{
				ResourceID: resource.ID,
				Resource:   resource,
				Role:       aa.Role,
				Options:    aa.Options,
			}
			appeal.SetDefaults()
			if err := s.providerService.ValidateAppeal(ctx, appeal, provider); err != nil {
				return fmt.Errorf("requirement[%v].appeals[%v]: %w", i, j, err)
			}
		}
	}
	return nil
}

func (s *Service) validateSteps(steps []*domain.Step) error {
	for _, step := range steps {
		if containsWhitespaces(step.Name) {
			return fmt.Errorf(`%w: "%s"`, ErrStepNameContainsWhitespaces, step.Name)
		}

		if step.Approvers != nil {
			for _, approver := range step.Approvers {
				if err := s.validateApprover(approver); err != nil {
					return fmt.Errorf(`validating approver "%s": %w`, approver, err)
				}
			}
		}
	}

	return nil
}

func (s *Service) validateApprover(expr string) error {
	if err := s.validator.Var(expr, "email"); err == nil {
		return nil
	}

	// skip validation if expression is accessing arbitrary value
	if strings.Contains(expr, "$appeal.resource.details") ||
		strings.Contains(expr, "$appeal.creator") {
		return nil
	}

	dummyAppeal := &domain.Appeal{
		Resource: &domain.Resource{},
	}
	dummyAppealMap, err := structToMap(dummyAppeal)
	if err != nil {
		return fmt.Errorf("parsing appeal to map: %w", err)
	}
	approvers, err := evaluator.Expression(expr).EvaluateWithVars(map[string]interface{}{
		"appeal": dummyAppealMap,
	})
	if err != nil {
		return fmt.Errorf("evaluating expression: %w", err)
	}

	// value type should be string or []string
	value := reflect.ValueOf(approvers)
	switch value.Type().Kind() {
	case reflect.String:
		return nil
	case reflect.Slice:
		elem := value.Type().Elem()
		switch elem.Kind() {
		case
			reflect.String,
			reflect.Interface: // can't determine exact type of interface{} elem
			return nil
		}
	}

	return fmt.Errorf(`invalid value type: "%s"`, expr)
}

func containsWhitespaces(s string) bool {
	r, _ := regexp.Compile(`\s`)
	return r.Match([]byte(s))
}

func structToMap(item interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if item != nil {
		jsonString, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jsonString, &result); err != nil {
			return nil, err
		}
	}

	return result, nil
}
