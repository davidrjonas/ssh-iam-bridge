package main

import (
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

func systemUsersAddToGroup(group string, names []string) {
	for _, name := range names {
		if err := exec.Command("usermod", "-a", "-G", group, name).Run(); err != nil {
			panic(err)
		}
	}
}

func systemUsersRemoveFromGroup(group string, names []string) {
	for _, name := range names {
		if err := exec.Command("gpasswd", "-d", name, group).Run(); err != nil {
			panic(err)
		}
	}
}

func systemUsersInGroup(name string) []string {

	out, err := exec.Command("getent", "group", name).Output()

	if err != nil {
		panic(err)
	}

	parts := strings.Split(strings.TrimSpace(string(out)), ":")
	users := strings.Split(parts[3], ",")

	return users
}

func createAsSystemGroup(name string, gid int) *user.Group {
	// Both exec.Command and user.LookupGroupId want a string
	id := strconv.Itoa(gid)

	if err := exec.Command("groupadd", "--gid", id, name).Run(); err != nil {
		panic(err)
	}

	g, _ := user.LookupGroupId(id)

	return g
}

func stringArrayContains(test string, set []string) bool {
	for _, member := range set {
		if test == member {
			return true
		}
	}

	return false
}

// Compares array1 against a2 and returns the values in array1 that are not present in a2.
func stringArrayDiff(tests, set []string) []string {
	missing := make([]string, 0)
	for _, test := range tests {
		if stringArrayContains(test, set) {
			continue
		}

		missing = append(missing, test)
	}

	return missing
}

func EnsureSystemGroup(group_name string, gid int, users []string) error {

	_, err := user.LookupGroup(group_name)

	if _, ok := err.(user.UnknownGroupError); !ok {
		createAsSystemGroup(group_name, gid)
	} else if err != nil {
		return err
	}

	system_users := systemUsersInGroup(group_name)

	systemUsersAddToGroup(group_name, stringArrayDiff(system_users, users))
	systemUsersRemoveFromGroup(group_name, stringArrayDiff(users, system_users))

	return nil
}
