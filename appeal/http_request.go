package appeal

import (
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
)

type resourceOptions struct {
	Role string `json:"role" validate:"required"`
}

type createPayloadResource struct {
	ID      uint                   `json:"id" validate:"required"`
	Options map[string]interface{} `json:"options"`
}

type createPayload struct {
	User      string                  `json:"user" validate:"required"`
	Resources []createPayloadResource `json:"resources" validate:"required,min=1"`
}

func (p *createPayload) toDomain() ([]*domain.Appeal, error) {
	appeals := []*domain.Appeal{}
	for _, r := range p.Resources {
		var options resourceOptions
		if err := mapstructure.Decode(r.Options, &options); err != nil {
			return nil, err
		}
		if err := utils.ValidateStruct(options); err != nil {
			return nil, err
		}

		appeals = append(appeals, &domain.Appeal{
			User:       p.User,
			ResourceID: r.ID,
			Role:       options.Role,
		})
	}

	return appeals, nil
}

type actionPayload struct {
	Actor  string `json:"actor"`
	Action string `json:"action"`
}

type revokePayload struct {
	Actor string `json:"actor"`
}
