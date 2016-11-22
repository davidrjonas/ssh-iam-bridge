package unix

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/davidrjonas/ssh-iam-bridge/string_array"
)

func addToGroup(group string, name string) {
	if err := exec.Command("usermod", "-a", "-G", group, name).Run(); err != nil {
		panic(err)
	}
}

func removeFromGroup(group string, name string) {
	if err := exec.Command("gpasswd", "-d", name, group).Run(); err != nil {
		panic(err)
	}
}

func usersInGroup(name string) []string {

	out, err := exec.Command("getent", "group", name).Output()

	if err != nil {
		panic(err)
	}

	parts := strings.Split(strings.TrimSpace(string(out)), ":")
	users := strings.Split(parts[3], ",")

	return users
}

func createGroup(name string, gid int) *user.Group {
	// Both exec.Command and user.LookupGroupId want a string
	id := strconv.Itoa(gid)

	if err := exec.Command("groupadd", "--gid", id, name).Run(); err != nil {
		panic(err)
	}

	g, _ := user.LookupGroupId(id)

	return g
}

func isManagedUser(min_user_id int) func(name string) bool {
	return func(name string) bool {
		return UserId(name) >= min_user_id
	}
}

func EnsureGroup(group_name string, gid int, users []string, min_user_id int) error {

	_, err := user.LookupGroup(group_name)

	if _, ok := err.(user.UnknownGroupError); ok {
		createGroup(group_name, gid)
	} else if err != nil {
		return err
	}

	users = string_array.Filter(users, UserExists)
	system_users := string_array.Filter(usersInGroup(group_name), isManagedUser(min_user_id))

	for _, name := range string_array.Diff(users, system_users) {
		fmt.Println("Adding", name, "to group", group_name)
		addToGroup(group_name, name)
	}

	for _, name := range string_array.Diff(system_users, users) {
		fmt.Println("Removing", name, "from group", group_name)
		removeFromGroup(group_name, name)
	}

	return nil
}

func UserExists(username string) bool {
	_, err := user.Lookup(username)

	if _, ok := err.(user.UnknownUserError); ok {
		return false
	} else if err != nil {
		panic(err)
	}

	return true
}

func UserId(username string) int {
	user, err := user.Lookup(username)

	if err != nil {
		return 0
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return 0
	}

	return uid
}

func EnsureUser(username string, uid int, comment string) error {
	if UserExists(username) {
		return nil
	}

	args := []string{
		"--create-home",
		"--user-group",
		"--shell", "/bin/bash",
		"--uid", strconv.Itoa(uid),
		"--comment", comment,
		username,
	}

	out, err := exec.Command("useradd", args...).Output()
	if err != nil {
		os.Stderr.Write(out)
		if exerr, ok := err.(*exec.ExitError); ok {
			os.Stderr.Write(exerr.Stderr)
		}
		panic(err)
	}

	return nil
}
