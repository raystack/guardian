package metabase

import (
	"fmt"
	"github.com/odpf/guardian/domain"
	. "github.com/odpf/guardian/plugins/migrations"
	"strconv"
)

type migration struct {
	typeName       string
	providerConfig domain.ProviderConfig
	resources      []domain.Resource
	pendingAppeals []domain.Appeal
}

const typeName = "metabase"

func NewMigration(providerConfig domain.ProviderConfig, resources []domain.Resource, pendingAppeals []domain.Appeal) *migration {
	return &migration{
		typeName:       typeName,
		providerConfig: providerConfig,
		resources:      resources,
		pendingAppeals: pendingAppeals,
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

	for _, appeal := range p.pendingAppeals {
		if m, ok := appealMap[appeal.ResourceID]; ok {
			appealMap[appeal.ResourceID] = append(m, appeal)
		} else {
			appealMap[appeal.ResourceID] = append(make([]domain.Appeal, 0), appeal)
		}
	}

	credentials := p.providerConfig.Credentials.(map[string]string)
	c, err := NewClient(&ClientConfig{
		Host:       credentials["host"],
		Username:   credentials["username"],
		Password:   credentials["password"],
		HTTPClient: nil,
	})

	userMap := make(map[int]user, 0)
	groupMap := make(map[int]group, 0)
	users, err := c.getUsers()

	for _, user := range users {
		userMap[user.ID] = user
	}

	groups, err := c.getGroups()
	for _, group := range groups {
		groupMap[group.ID] = group
	}

	membership, err := c.getMembership()
	appeals := make([]AppealRequest, 0)
	for userID, members := range membership {
		for _, m := range members {
			userIdInt, _ := strconv.Atoi(userID)
			if user, ok := userMap[userIdInt]; ok {
				resourceUrn := fmt.Sprintf("group:%d", m.GroupId)
				if resource, ok := resourceMap[resourceUrn]; ok {
					if _, ok := appealMap[resource.ID]; !ok {
						appeal := AppealRequest{
							AccountID: user.Email,
							User:      user.Email,
							Resource:  ResourceRequest{ID: resource.ID, Duration: "30d"},
						}
						appeals = append(appeals, appeal)
					}
				}
			}

		}
	}
	return appeals, err
}
