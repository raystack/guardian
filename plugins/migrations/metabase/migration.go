package metabase

import (
	"fmt"
	"strconv"

	"github.com/odpf/guardian/domain"
	. "github.com/odpf/guardian/plugins/migrations"
)

type migration struct {
	typeName        string
	providerConfig  *domain.ProviderConfig
	resources       []domain.Resource
	excludedAppeals []domain.Appeal
}

const typeName = Metabase

func NewMigration(providerConfig *domain.ProviderConfig, resources []domain.Resource, excludedAppeals []domain.Appeal) *migration {
	return &migration{
		typeName:        typeName,
		providerConfig:  providerConfig,
		resources:       resources,
		excludedAppeals: excludedAppeals,
	}
}

func (p *migration) GetType() string {
	return p.typeName
}

func (p *migration) PopulateAccess() ([]AppealRequest, error) {
	resourceMap := make(map[string]domain.Resource, 0)
	appealMap := make(map[string][]domain.Appeal, 0)

	for _, resource := range p.resources {
		resourceMap[resource.URN] = resource
	}

	for _, appeal := range p.excludedAppeals {
		if m, ok := appealMap[appeal.ResourceID]; ok {
			appealMap[appeal.ResourceID] = append(m, appeal)
		} else {
			appealMap[appeal.ResourceID] = append(make([]domain.Appeal, 0), appeal)
		}
	}

	credentials := p.providerConfig.Credentials.(map[string]string)
	c, err := NewClient(&ClientConfig{
		Host:       credentials[Host],
		Username:   credentials[Username],
		Password:   credentials[Password],
		HTTPClient: nil,
	})
	if err != nil {
		return nil, err
	}

	userMap := make(map[int]user, 0)
	users, err := c.getUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		userMap[user.ID] = user
	}

	membership, err := c.getMembership()
	if err != nil {
		return nil, err
	}

	appeals := make([]AppealRequest, 0)
	for userID, members := range membership {
		for _, m := range members {
			userIdInt, _ := strconv.Atoi(userID)
			if user, ok := userMap[userIdInt]; ok {
				resourceUrn := fmt.Sprintf("%s:%d", Group, m.GroupId)
				if resource, ok := resourceMap[resourceUrn]; ok {
					if _, ok := appealMap[resource.ID]; !ok {
						appeal := AppealRequest{
							AccountID: user.Email,
							User:      user.Email,
							Resource:  ResourceRequest{ID: resource.ID, Name: resource.Name, Role: Member, Duration: DefaultDuration},
						}
						appeals = append(appeals, appeal)
					}
				}
			}
		}
	}
	return appeals, err
}
