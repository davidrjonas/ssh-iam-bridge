package main

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

var iam_svc *iam.IAM

func getIamService() *iam.IAM {
	if iam_svc != nil {
		return iam_svc
	}

	sess, err := session.NewSession()

	if err != nil {
		panic(err)
	}

	iam_svc := iam.New(sess, &aws.Config{Region: aws.String("us-east-1")})

	return iam_svc
}

func isPrefixedBy(s *string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(*s, prefix) {
			return true
		}
	}
	return false
}

func filterGroups(groups []*iam.Group, cb func(*iam.Group) bool) []*iam.Group {

	filtered := make([]*iam.Group, 0)

	for _, g := range groups {
		if cb(g) {
			filtered = append(filtered, g)
		}
	}

	return filtered
}

func GetGroups(prefixes []string) ([]*iam.Group, error) {

	svc := getIamService()

	resp, err := svc.ListGroups(nil)

	if err != nil {
		return []*iam.Group{}, err
	}

	return filterGroups(resp.Groups, func(g *iam.Group) bool {
		return isPrefixedBy(g.GroupName, prefixes)
	}), nil
}

func GetGroupUsers(group *iam.Group) ([]*iam.User, error) {

	svc := getIamService()

	resp, err := svc.GetGroup(&iam.GetGroupInput{GroupName: group.GroupName})

	if err != nil {
		return []*iam.User{}, err
	}

	return resp.Users, nil
}

func ensureSystemGroup(group *iam.Group, users []*iam.User) error {
	// Create system group if it doesn' exist
	_, err := user.LookupGroup(aws.StringValue(group.GroupName))

	if _, ok := err.(user.UnknownGroupError); !ok {
		// create group
	} else if err != nil {
		return err
	}

	// TODO: Get the users in the group. Determine which we should add / remove
	return nil
}

func SyncGroups(prefix string) error {

	role, err := GetIamRole()
	if err != nil {
		// FIXME: Just log it.
	}

	var prefixes []string

	if role != "" {
		prefixes = []string{prefix, fmt.Sprintf("%s%s-", prefix, role)}
	} else {
		prefixes = []string{prefix}
	}

	groups, err := GetGroups(prefixes)
	if err != nil {
		return err
	}

	for _, group := range groups {
		users, err := GetGroupUsers(group)
		if err != nil {
			// FIXME: log it
			continue
		}

		err = ensureSystemGroup(group, users)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetAuthorizedKeys(username string) (*bytes.Buffer, error) {

	svc := getIamService()

	resp, err := svc.ListSSHPublicKeys(&iam.ListSSHPublicKeysInput{UserName: &username})

	if err != nil {
		return nil, err
	}

	var out bytes.Buffer

	for _, metaref := range resp.SSHPublicKeys {
		if *metaref.Status != iam.StatusTypeActive {
			continue
		}

		keyref, err := svc.GetSSHPublicKey(&iam.GetSSHPublicKeyInput{
			SSHPublicKeyId: metaref.SSHPublicKeyId,
			UserName:       metaref.UserName,
			Encoding:       aws.String(iam.EncodingTypeSsh),
		})

		if err != nil {
			return nil, err
		}

		fmt.Fprintln(&out, "# Key id: ", *metaref.SSHPublicKeyId)
		fmt.Fprintln(&out, *keyref.SSHPublicKey.SSHPublicKeyBody)
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
