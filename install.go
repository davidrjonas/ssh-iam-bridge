package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"

	"github.com/davidrjonas/ssh-iam-bridge/string_array"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func backupFile(filename string) {
	fmt.Println("Backing up", filename, "to", filename+".orig")
	err := exec.Command("cp", "-f", filename, filename+".orig").Run()
	check(err)
}

func install(self_path, username string) {
	cmd_name := installAuthorizedKeysCommandScript(self_path)
	installUser(username)
	installToSshd(cmd_name, username)
	installToPam(self_path)
	installToCron(self_path)
}

// ssh is picky about AuthorizedKeysCommand, see man sshd_config
func installAuthorizedKeysCommandScript(self_path string) string {
	cmd_name := "/usr/sbin/ssh-iam-bridge-public-keys"
	fmt.Println("Writing AuthorizedKeysCommand script", cmd_name)

	script := fmt.Sprintf("#!/bin/sh\nexec %s authorized_keys \"$@\"\n", self_path)

	check(ioutil.WriteFile(cmd_name, []byte(script), 0755))

	return cmd_name
}

func installUser(username string) {
	_, err := user.Lookup(username)

	if err == nil {
		// User already exists
		return
	}

	if _, ok := err.(user.UnknownUserError); !ok {
		panic(err)
	}

	fmt.Println("Creating SSH authorized keys lookup user", username)

	args := []string{
		"--system",
		"--shell", "/usr/sbin/nologin",
		"--comment", "SSH authorized keys lookup",
		username,
	}

	check(exec.Command("useradd", args...).Run())
}

func installToSshd(cmd, username string) {

	filename := "/etc/ssh/sshd_config"

	lines_to_add := []string{
		"AuthorizedKeysCommand " + cmd + "\n",
		"AuthorizedKeysCommandUser " + username + "\n",
		"ChallengeResponseAuthentication yes\n",
		"AuthenticationMethods publickey keyboard-interactive:pam,publickey\n",
	}

	lines := string_array.ReadFile(filename)

	if string_array.ContainsAll(lines, lines_to_add) {
		return
	}

	fmt.Println("Updating", filename)

	// Comment out specific lines
	lines_to_comment := []string{
		"AuthorizedKeysCommand ",
		"AuthorizedKeysCommandUser ",
		"ChallengeResponseAuthentication ",
		"AuthenticationMethods ",
	}

	for idx, line := range lines {
		for _, check := range lines_to_comment {
			if strings.HasPrefix(line, check) {
				lines[idx] = "# " + line
			}
		}
	}

	backupFile(filename)

	check(string_array.WriteFile(filename, lines, lines_to_add))

	err := exec.Command("sshd", "-t").Run()

	if err != nil {
		if exerr, ok := err.(*exec.ExitError); ok {
			os.Stderr.Write(exerr.Stderr)
			os.Exit(1)
		}

		panic(err)
	}
}

func installToPam(self_path string) {

	filename := "/etc/pam.d/sshd"
	fmt.Println("Updating", filename)

	pam_exec := "auth requisite pam_exec.so stdout " + self_path + " pam_create_user\n"

	lines := string_array.ReadFile(filename)

	for _, line := range lines {
		if line == pam_exec {
			return
		}
	}

	backupFile(filename)
	check(string_array.WriteFile(filename, []string{"# Next line added by " + self_path + "\n", pam_exec}, lines))
}

func installToCron(self_path string) {

	filename := "/etc/cron.d/" + path.Base(self_path)

	fmt.Println("Installing crontab", filename)

	contents := "*/10 * * * * root " + self_path + " sync_groups"

	check(ioutil.WriteFile(filename, []byte(contents), 0644))
}
