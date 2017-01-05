package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/davidrjonas/ssh-iam-bridge/directory"
	"github.com/davidrjonas/ssh-iam-bridge/strarray"
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
			cg.Users = strarray.Unique(append(cg.Users, usernames...))
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
	return unix.UserID(name) >= uidOffset
}

func createUserFromName(name string) error {
	user, err := directory.GetUser(name)

	if err != nil {
		return err
	}

	if err = unix.EnsureUser(name, awsToUnixID(user.UserId), "iam="+aws.StringValue(user.UserId)); err != nil {
		return err
	}

	return nil
}

func ensureGroup(name string, gid int, users []string, ignoreMissingUsers bool) error {
	if err := unix.EnsureGroup(name, gid); err != nil {
		return err
	}

	systemUsers, err := unix.UsersInGroup(name)
	if err != nil {
		return err
	}

	if ignoreMissingUsers {
		users = strarray.Filter(users, unix.UserExists)
	} else {
		missing := strarray.Filter(users, func(name string) bool { return !unix.UserExists(name) })
		for _, username := range missing {
			createUserFromName(username)
		}
	}

	systemUsers = strarray.Filter(systemUsers, isManagedUser)

	for _, username := range strarray.Diff(users, systemUsers) {
		fmt.Println("Adding", username, "to group", name)
		if err = unix.AddToGroup(name, username); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add user '%s' to group '%s'; %s", username, name, err)
			os.Exit(1)
		}
	}

	for _, username := range strarray.Diff(systemUsers, users) {
		fmt.Println("Removing", username, "from group", name)
		if err = unix.RemoveFromGroup(name, username); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove user '%s' from group '%s'; %s", username, name, err)
			os.Exit(1)
		}
	}

	return nil
}

func syncGroups(prefix string) error {
	return sync(prefix, true)
}

func sync(prefix string, ignoreMissingUsers bool) error {

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
		if err := ensureGroup(name, groupIDForGroups(cg.Sources), cg.Users, ignoreMissingUsers); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add group '%s'; %s", name, err)
			os.Exit(1)
		}
	}

	return nil
}
