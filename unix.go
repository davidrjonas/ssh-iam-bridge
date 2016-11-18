package main

import (
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/davidrjonas/ssh-iam-bridge/string_array"
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

func EnsureSystemGroup(group_name string, gid int, users []string) error {

	_, err := user.LookupGroup(group_name)

	if _, ok := err.(user.UnknownGroupError); ok {
		createAsSystemGroup(group_name, gid)
	} else if err != nil {
		return err
	}

	system_users := systemUsersInGroup(group_name)

	systemUsersAddToGroup(group_name, string_array.Diff(system_users, users))
	systemUsersRemoveFromGroup(group_name, string_array.Diff(users, system_users))

	return nil
}

func userExists(username string) bool {
	_, err := user.Lookup(username)

	if _, ok := err.(user.UnknownUserError); ok {
		return false
	} else if err != nil {
		panic(err)
	}

	return true
}

func EnsureUser(username string, uid int, comment string) error {
	if userExists(username) {
		return nil
	}

	cmd := []string{
		"useradd",
		"--create-home",
		"--user-group",
		"--shell", "/bin/bash",
		"--uid", uid,
		"--comment", comment,
		username
	}

	// TODO: wrap error
	return exec.Command(cmd...).Run()
}
