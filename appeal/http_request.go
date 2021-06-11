package appeal

import (
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
)

type resourceOptions struct {
	Duration string `json:"duration"`
}

type createPayloadResource struct {
	ID      uint                   `json:"id" validate:"required"`
	Role    string                 `json:"role" validate:"required"`
	Options map[string]interface{} `json:"options"`
}

type createPayload struct {
	User      string                  `json:"user" validate:"required"`
	Resources []createPayloadResource `json:"resources" validate:"required,min=1"`
}

func (p *createPayload) toDomain() ([]*domain.Appeal, error) {
	appeals := []*domain.Appeal{}
	for _, r := range p.Resources {
		var options *domain.AppealOptions

		var resOptions *resourceOptions
		if err := mapstructure.Decode(r.Options, &resOptions); err != nil {
			return nil, err
		}
		if resOptions != nil {
			if err := utils.ValidateStruct(resOptions); err != nil {
				return nil, err
			}
			var expirationDate time.Time
			if resOptions.Duration != "" {
				duration, err := time.ParseDuration(resOptions.Duration)
				if err != nil {
					return nil, err
				}
				expirationDate = TimeNow().Add(duration)
			}

			options = &domain.AppealOptions{
				ExpirationDate: &expirationDate,
			}
		}

		appeals = append(appeals, &domain.Appeal{
			User:       p.User,
			ResourceID: r.ID,
			Role:       r.Role,
			Options:    options,
		})
	}

	return appeals, nil
}

type actionPayload struct {
	Action string `json:"action"`
}
