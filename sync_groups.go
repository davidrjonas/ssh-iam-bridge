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

type combinedGroup struct {
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

func groupIDForGroups(groups []*iam.Group) int {

	if len(groups) == 1 {
		return awsToUnixID(groups[0].GroupId)
	}

	// Use the shortest group name as the group id
	var minGroup *iam.Group

	min := 9999
	for _, group := range groups {
		l := len(aws.StringValue(group.GroupName))
		if l < min {
			min = l
			minGroup = group
		}
	}

	return awsToUnixID(minGroup.GroupId)
}

func coalesceGroups(groups []*iam.Group, prefixes []string) map[string]combinedGroup {
	combined := make(map[string]combinedGroup, 0)

	for _, group := range groups {
		users, err := directory.GetGroupUsers(group)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get IAM group users for group %s (continuing), %s", aws.StringValue(group.GroupName), err)
			continue
		}

		usernames := directory.UserNames(users)

		name, _ := removePrefix(aws.StringValue(group.GroupName), prefixes)

		if _, ok := combined[name]; ok {
			cg := combined[name]
			cg.Sources = append(cg.Sources, group)
			cg.Users = string_array.Unique(append(cg.Users, usernames...))
		} else {
			combined[name] = combinedGroup{
				Sources: []*iam.Group{group},
				Users:   usernames,
			}
		}
	}

	return combined
}

func isManagedUser(name string) bool {
	return unix.UserId(name) >= UID_OFFSET
}

func ensureGroup(name string, gid int, users []string) error {
	if err := unix.EnsureGroup(name, gid); err != nil {
		return err
	}

	users = string_array.Filter(users, unix.UserExists)
	systemUsers := string_array.Filter(unix.UsersInGroup(name), isManagedUser)

	for _, username := range string_array.Diff(users, systemUsers) {
		fmt.Println("Adding", username, "to group", name)
		unix.AddToGroup(name, username)
	}

	for _, username := range string_array.Diff(systemUsers, users) {
		fmt.Println("Removing", username, "from group", name)
		unix.RemoveFromGroup(name, username)
	}

	return nil
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

	for name, cg := range coalesceGroups(groups, prefixes) {
		if err := ensureGroup(name, groupIDForGroups(cg.Sources), cg.Users); err != nil {
			panic(err)
		}
	}

	return nil
}
