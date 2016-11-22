package unix

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

func AddToGroup(group string, name string) {
	if err := exec.Command("usermod", "-a", "-G", group, name).Run(); err != nil {
		panic(err)
	}
}

func RemoveFromGroup(group string, name string) {
	if err := exec.Command("gpasswd", "-d", name, group).Run(); err != nil {
		panic(err)
	}
}

func UsersInGroup(name string) []string {

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

func EnsureGroup(groupName string, gid int) error {

	_, err := user.LookupGroup(groupName)

	if _, ok := err.(user.UnknownGroupError); ok {
		createGroup(groupName, gid)
	} else if err != nil {
		return err
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

func UserID(username string) int {
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
