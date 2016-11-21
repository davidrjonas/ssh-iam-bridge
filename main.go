package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/davidrjonas/ssh-iam-bridge/unix"
	"github.com/kardianos/osext"
)

const NAME = "ssh-iam-bridge"
const VERSION = "1.0.0"
const PREFIX = "server-"

const EX_EINVAL = 22
const EX_TEMPFAIL = 75
const EX_NOPERM = 77

const UID_OFFSET = 2000

func getPrefix() string {
	return PREFIX
}

func awsToUnixId(aws_id *string) int {
	// Treat the last 2 bytes of a sha256 hash of aws_id as an uint and add it to 2000
	b := []byte(*aws_id)

	hasher := sha256.New()
	hasher.Write(b)
	h := hasher.Sum(nil)

	data, _ := binary.ReadUvarint(bytes.NewBuffer(h[len(h)-2:]))

	return UID_OFFSET + (int(data) / 2)
}

func GetAuthorizedKeys(username string) (*bytes.Buffer, error) {

	keys, err := GetActiveSshPublicKeys(username)

	if err != nil {
		return nil, err
	}

	var out bytes.Buffer

	for _, key := range keys {

		body, err := GetSshEncodePublicKey(key.UserName, key.SSHPublicKeyId)

		if err != nil {
			return nil, err
		}

		fmt.Fprintln(&out, "# Key id: ", *key.SSHPublicKeyId)
		fmt.Fprintln(&out, *body)
	}

	return &out, nil
}

func printAuthorizedKeys(username string) error {
	buf, err := GetAuthorizedKeys(username)
	if err != nil {
		return err
	}
	buf.WriteTo(os.Stdout)
	return nil
}

func pamCreateUser() {
	username := os.Getenv("PAM_USER")

	if username == "" {
		os.Stderr.WriteString("Unable to find pam user in the environment\n")
		os.Exit(EX_EINVAL)
	}

	// PATH is not set when called from pam.
	os.Setenv("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")

	if unix.UserExists(username) {
		// Supposedly: Terminate the PAM authentication stack. The SSH client
		// will fail since the user didn't supply a valid public key.
		os.Exit(EX_NOPERM)
	}

	user, err := GetUser(username)

	if err != nil {
		panic(err)
	}

	err = unix.EnsureUser(username, awsToUnixId(user.UserId), "iam="+aws.StringValue(user.UserId))

	if err != nil {
		panic(err)
	}

	syncGroups(getPrefix())

	fmt.Println(NAME + ": Your user has been created but you must reconnect to for it to be active.")
	fmt.Println(NAME + ": Connect again to log in to your account.")

	os.Exit(EX_TEMPFAIL)
}

func getSelfPath() string {
	self, err := osext.Executable()
	if err != nil {
		panic(err)
	}
	return self
}

var (
	installCommand       = kingpin.Command("install", "Install this program to authenticate SSH connections and create users")
	installCommandUser   = installCommand.Arg("user", "The user under which to run the AuthorizedKeysCommand, will be created if it doesn't exit").Default("ssh-iam-bridge").String()
	authKeysCommand      = kingpin.Command("authorized_keys", "Get the authorized_keys from IAM for user")
	authKeysCommandUser  = authKeysCommand.Arg("user", "The IAM username for which to get keys").Required().String()
	syncGroupsCommand    = kingpin.Command("sync_groups", "Sync the IAM groups with the local system groups")
	pamCreateUserCommand = kingpin.Command("pam_create_user", "Create a user from the env during the sshd pam phase")
)

func main() {

	kingpin.Version(VERSION)

	switch kingpin.Parse() {
	case installCommand.FullCommand():
		install(getSelfPath(), *installCommandUser)
	case authKeysCommand.FullCommand():
		printAuthorizedKeys(*authKeysCommandUser)
	case syncGroupsCommand.FullCommand():
		syncGroups(getPrefix())
	case pamCreateUserCommand.FullCommand():
		pamCreateUser()
	}
}
