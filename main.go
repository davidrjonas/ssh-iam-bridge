package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/davidrjonas/ssh-iam-bridge/directory"
	"github.com/davidrjonas/ssh-iam-bridge/unix"
	"github.com/kardianos/osext"
)

var version = "1.0.0"

const exitEInval = 22
const exitTempFail = 75
const exitNoPerm = 77

const uidOffset = 2000

func getName() string {
	return "ssh-iam-bridge"
}

func getPrefix() string {
	return "server-"
}

func awsToUnixID(awsID *string) int {
	// Treat the last 2 bytes of a sha256 hash of awsId as an uint and add it to 2000
	b := []byte(*awsID)

	hasher := sha256.New()
	hasher.Write(b)
	h := hasher.Sum(nil)

	data, _ := binary.ReadUvarint(bytes.NewBuffer(h[len(h)-2:]))

	return uidOffset + (int(data) / 2)
}

func getAuthorizedKeys(username string) (*bytes.Buffer, error) {

	keys, err := directory.GetActiveSshPublicKeys(username)

	if err != nil {
		return nil, err
	}

	var out bytes.Buffer

	for _, key := range keys {

		body, err := directory.GetSshEncodedPublicKey(key.UserName, key.SSHPublicKeyId)

		if err != nil {
			return nil, err
		}

		fmt.Fprintln(&out, "# Key id: ", *key.SSHPublicKeyId)
		fmt.Fprintln(&out, *body)
	}

	return &out, nil
}

func printAuthorizedKeys(username string) error {
	buf, err := getAuthorizedKeys(username)
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
		os.Exit(exitEInval)
	}

	// PATH is not set when called from pam.
	os.Setenv("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")

	if unix.UserExists(username) {
		// Supposedly: Terminate the PAM authentication stack. The SSH client
		// will fail since the user didn't supply a valid public key.
		os.Exit(exitNoPerm)
	}

	user, err := directory.GetUser(username)

	if err != nil {
		panic(err)
	}

	err = unix.EnsureUser(username, awsToUnixID(user.UserId), "iam="+aws.StringValue(user.UserId))

	if err != nil {
		panic(err)
	}

	syncGroups(getPrefix())

	fmt.Println(getName() + ": Your user has been created but you must reconnect to for it to be active.")
	fmt.Println(getName() + ": Connect again to log in to your account.")

	os.Exit(exitTempFail)
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

	kingpin.Version(version)

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
