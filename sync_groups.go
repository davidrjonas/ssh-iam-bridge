package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/davidrjonas/ssh-iam-bridge/directory"
	"github.com/davidrjonas/ssh-iam-bridge/string_array"
	"github.com/davidrjonas/ssh-iam-bridge/unix"
)

type CombinedGroup struct {
	Sources []*iam.Group
	Users   []string
}

func removePrefix(s string, prefixes []string) (string, string) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return strings.TrimPrefix(s, prefix), prefix
		}
	}
	return s, ""
}

func syncGroups(prefix string) error {

	role, err := directory.GetRole()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting IAM role (continuing), %s", err)
	}

	var prefixes []string

	if role != "" {
		prefixes = []string{prefix, prefix + role + "-"}
	} else {
		prefixes = []string{prefix}
	}

	groups, err := directory.GetGroups(prefixes)

	if err != nil {
		return err
	}

	// Coalesce
	var group_map map[string]CombinedGroup

	for _, group := range groups {
		users, err := directory.GetGroupUsers(group)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get IAM group users for group %s (continuing), %s", aws.StringValue(group.GroupName), err)
			continue
		}

		usernames := directory.UserNames(users)

		name, _ := removePrefix(aws.StringValue(group.GroupName), prefixes)

		if _, ok := group_map[name]; ok {
			cg := group_map[name]
			cg.Sources = append(cg.Sources, group)
			cg.Users = string_array.Unique(append(cg.Users, usernames...))
		} else {
			group_map[name] = CombinedGroup{
				Sources: []*iam.Group{group},
				Users:   usernames,
			}
		}
	}

	for name, cg := range group_map {
		var gid int

		if len(cg.Sources) == 1 {
			gid = awsToUnixId(cg.Sources[0].GroupId)
		} else {
			// Use the shortest group name as the group id
			var source *iam.Group

			min := 9999
			for _, g := range cg.Sources {
				l := len(aws.StringValue(g.GroupName))
				if l < min {
					min = l
					source = g
				}
			}

			gid = awsToUnixId(source.GroupId)
		}

		err = unix.EnsureGroup(name, gid, cg.Users, UID_OFFSET)

		if err != nil {
			return err
		}
	}

	return nil
}
