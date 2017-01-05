package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"

	"github.com/davidrjonas/ssh-iam-bridge/strarray"
)

func check(err error, failureMessage string) {
	if err == nil {
		return
	}

	if exerr, ok := err.(*exec.ExitError); ok {
		os.Stderr.Write(exerr.Stderr)
	}

	fmt.Fprintln(os.Stderr, "!!!", failureMessage)
	os.Exit(1)
}

func backupFile(filename string) {
	fmt.Println("Backing up", filename, "to", filename+".orig")
	err := exec.Command("cp", "-f", filename, filename+".orig").Run()
	check(err, "Failed to backup "+filename)
}

func install(selfPath, username string, noPam bool) {
	cmdName := installAuthorizedKeysCommandScript(selfPath)
	installUser(username)
	installToSshd(cmdName, username)
	if !noPam {
		installToPam(selfPath)
	}
	installToCron(selfPath, noPam)

	fmt.Println("* Restart sshd for changes to take effect")
}

// ssh is picky about AuthorizedKeysCommand, see man sshd_config
func installAuthorizedKeysCommandScript(selfPath string) string {
	cmdName := "/usr/sbin/ssh-iam-bridge-public-keys"
	fmt.Println("Writing AuthorizedKeysCommand script", cmdName)

	script := fmt.Sprintf("#!/bin/sh\nexec %s authorized_keys \"$@\"\n", selfPath)

	check(ioutil.WriteFile(cmdName, []byte(script), 0755), "Failed to write script"+cmdName)

	return cmdName
}

func installUser(username string) {
	_, err := user.Lookup(username)

	if err == nil {
		// User already exists
		return
	}

	if _, ok := err.(user.UnknownUserError); !ok {
		fmt.Fprintf(os.Stderr, "Failed to lookup user '%s'; %s", username, err)
		os.Exit(1)
	}

	fmt.Println("Creating SSH authorized keys lookup user", username)

	args := []string{
		"--system",
		"--shell", "/usr/sbin/nologin",
		"--comment", "SSH authorized keys lookup",
		username,
	}

	_, err = exec.Command("useradd", args...).Output()
	check(err, "Failed to create authorized keys lookup user")
}

func installToSshd(cmd, username string) {

	filename := "/etc/ssh/sshd_config"

	// TODO: Ensure 'PasswordAuthentication no' and 'UsePAM yes'

	linesToAdd := []string{
		"AuthorizedKeysCommand " + cmd + "\n",
		"AuthorizedKeysCommandUser " + username + "\n",
		"ChallengeResponseAuthentication yes\n",
		"AuthenticationMethods publickey keyboard-interactive:pam,publickey\n",
	}

	lines, err := strarray.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read '%s'; %s", filename, err)
		os.Exit(1)
	}

	if strarray.ContainsAll(lines, linesToAdd) {
		return
	}

	fmt.Println("Updating", filename)

	// Comment out specific lines
	linesToComment := []string{
		"AuthorizedKeysCommand ",
		"AuthorizedKeysCommandUser ",
		"ChallengeResponseAuthentication ",
		"AuthenticationMethods ",
	}

	for idx, line := range lines {
		for _, check := range linesToComment {
			if strings.HasPrefix(line, check) {
				lines[idx] = "# " + line
			}
		}
	}

	backupFile(filename)

	check(strarray.WriteFile(filename, lines, linesToAdd), "Failed to write changes to "+filename)

	err = exec.Command("sshd", "-t").Run()
	check(err, "Modified sshd_config failed to lint. Correct the errors before proceeding.")
}

func installToPam(selfPath string) {

	filename := "/etc/pam.d/sshd"
	fmt.Println("Updating", filename)

	pamExec := "auth requisite pam_exec.so stdout quiet " + selfPath + " pam_create_user\n"

	lines, err := strarray.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read '%s'; %s", filename, err)
	}

	for _, line := range lines {
		if line == pamExec {
			return
		}
	}

	backupFile(filename)
	check(strarray.WriteFile(filename, []string{"# Next line added by " + selfPath + "\n", pamExec}, lines), "Failed to modify pam")
}

func installToCron(selfPath string, createMissingUsers bool) {

	options := []string{}

	if !createMissingUsers {
		options = append(options, "--ignore-missing-users")
	}

	filename := "/etc/cron.d/" + path.Base(selfPath)

	fmt.Println("Installing crontab", filename)

	contents := "*/10 * * * * root " + selfPath + " sync " + strings.Join(options, " ")

	check(ioutil.WriteFile(filename, []byte(contents), 0644), "Failed to write "+filename)
}
