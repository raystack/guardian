package newpoc

import (
	"fmt"
	"strings"
)

func containsString(arr []string, v string) bool {
	for _, item := range arr {
		if item == v {
			return true
		}
	}
	return false
}

func getResourceIdentifier(urn string) (rType, id string, err error) {
	resourceName := strings.Split(urn, "/")
	if len(resourceName) != 2 {
		return "", "", fmt.Errorf("invalid resource name: %s", urn)
	}
	resourceType := resourceName[0]
	if resourceType == "projects" {
		resourceType = ResourceTypeProject
	} else if resourceType == "organizations" {
		resourceType = ResourceTypeOrganization
	}

	return resourceType, resourceName[1], nil
}
