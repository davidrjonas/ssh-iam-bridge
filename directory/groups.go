package directory

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/iam"
)

func isPrefixedBy(s *string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(*s, prefix) {
			return true
		}
	}
	return false
}

func filterGroups(groups []*iam.Group, cb func(*iam.Group) bool) []*iam.Group {

	filtered := make([]*iam.Group, 0)

	for _, g := range groups {
		if cb(g) {
			filtered = append(filtered, g)
		}
	}

	return filtered
}

func GetGroups(prefixes []string) ([]*iam.Group, error) {

	svc := getIamService()

	resp, err := svc.ListGroups(nil)

	if err != nil {
		return []*iam.Group{}, err
	}

	return filterGroups(resp.Groups, func(g *iam.Group) bool {
		return isPrefixedBy(g.GroupName, prefixes)
	}), nil
}

func GetGroupUsers(group *iam.Group) ([]*iam.User, error) {

	svc := getIamService()

	resp, err := svc.GetGroup(&iam.GetGroupInput{GroupName: group.GroupName})

	if err != nil {
		return []*iam.User{}, err
	}

	return resp.Users, nil
}
