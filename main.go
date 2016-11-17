package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
)

func awsToUnixId(aws_id *string) int {
	// Treat the last 2 bytes of a sha256 hash of aws_id as an uint and add it to 2000
	b := []byte(*aws_id)

	hasher := sha256.New()
	hasher.Write(b)
	h := hasher.Sum(nil)

	data, _ := binary.ReadUvarint(bytes.NewBuffer(h[len(h)-2:]))

	return 2000 + (int(data) / 2)
}

func SyncGroups(prefix string) error {

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

		err = EnsureSystemGroup(*group.GroupName, awsToUnixId(group.GroupId), iamUserNames(users))
		if err != nil {
			return err
		}
	}

	return nil
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

func main() {
	buf, err := GetAuthorizedKeys("djonas")
	if err != nil {
		panic(err)
	}
	buf.WriteTo(os.Stdout)
}
