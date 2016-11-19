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

const VERSION = "1.0.0"
const PREFIX = "system-"

func awsToUnixId(aws_id *string) int {
	// Treat the last 2 bytes of a sha256 hash of aws_id as an uint and add it to 2000
	b := []byte(*aws_id)

	hasher := sha256.New()
	hasher.Write(b)
	h := hasher.Sum(nil)

	data, _ := binary.ReadUvarint(bytes.NewBuffer(h[len(h)-2:]))

	return 2000 + (int(data) / 2)
}

func syncGroups(prefix string) error {

	role, err := GetIamRole()

	if err != nil {
		// FIXME: Just log it.
	}

	var prefixes []string

	if role != "" {
		prefixes = []string{prefix, prefix + role + "-"}
	} else {
		prefixes = []string{prefix}
	}

	groups, err := GetIamGroups(prefixes)

	if err != nil {
		return err
	}

	for _, group := range groups {
		users, err := GetIamGroupUsers(group)
		if err != nil {
			// FIXME: log it
			continue
		}

		err = unix.EnsureGroup(*group.GroupName, awsToUnixId(group.GroupId), iamUserNames(users))
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateUser(username string) error {
	user, err := GetUser(username)

	if err != nil {
		return err
	}

	return unix.EnsureUser(username, awsToUnixId(user.UserId), aws.StringValue(user.Arn))
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

func pamCreateUser() error {
	return nil
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
	installCommandUser   = installCommand.Arg("user", "The user under which to run the AuthorizedKeysCommand, will be created if it doesn't exit").String()
	authKeysCommand      = kingpin.Command("authorized_keys", "Get the authorized_keys from IAM for user")
	authKeysCommandUser  = authKeysCommand.Arg("user", "The IAM username for which to get keys").Required().String()
	syncGroupsCommand    = kingpin.Command("sync_groups", "Sync the IAM groups with the local system groups")
	pamCreateUserCommand = kingpin.Command("pam_create_user", "Create a user from the env during the sshd pam phase")
	testCommand          = kingpin.Command("test", "")
)

func runTest() error {
	fmt.Println(getSelfPath())
	return nil
}

func main() {

	kingpin.Version(VERSION)

	switch kingpin.Parse() {
	case installCommand.FullCommand():
		install(getSelfPath(), *installCommandUser)
	case authKeysCommand.FullCommand():
		printAuthorizedKeys(*authKeysCommandUser)
	case syncGroupsCommand.FullCommand():
		syncGroups(PREFIX)
	case pamCreateUserCommand.FullCommand():
		pamCreateUser()
	case testCommand.FullCommand():
		runTest()
	}
}
