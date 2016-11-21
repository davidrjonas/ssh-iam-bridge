package unix

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/davidrjonas/ssh-iam-bridge/string_array"
)

func addToGroup(group string, names []string) {
	for _, name := range names {
		if !UserExists(name) {
			continue
		}
		if err := exec.Command("usermod", "-a", "-G", group, name).Run(); err != nil {
			panic(err)
		}
	}
}

func removeFromGroup(group string, names []string, min_user_id int) {
	for _, name := range names {
		// don't remove users with ids lower than our min_user_id
		if UserId(name) < min_user_id {
			continue
		}

		if err := exec.Command("gpasswd", "-d", name, group).Run(); err != nil {
			panic(err)
		}
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

func EnsureGroup(group_name string, gid int, users []string, min_user_id int) error {

	_, err := user.LookupGroup(group_name)

	if _, ok := err.(user.UnknownGroupError); ok {
		createGroup(group_name, gid)
	} else if err != nil {
		return err
	}

	system_users := usersInGroup(group_name)

	addToGroup(group_name, string_array.Diff(system_users, users))
	removeFromGroup(group_name, string_array.Diff(users, system_users), min_user_id)

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
