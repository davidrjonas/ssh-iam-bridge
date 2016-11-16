package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

func isPrefixedBy(s *string, prefixes []string) bool {
	return true
}

func GetGroups(prefixes []string) ([]*iam.Group, error) {

	sess, err := session.NewSession()

	if err != nil {
		return []*iam.Group{}, err
	}

	svc := iam.New(sess, &aws.Config{Region: aws.String("us-east-1")})

	resp, err := svc.ListGroups(nil)

	if err != nil {
		return []*iam.Group{}, err
	}

	groups := make([]*iam.Group, 0)

	for _, g := range resp.Groups {
		if isPrefixedBy(g.GroupName, prefixes) {
			groups = append(groups, g)
		}
	}

	return groups, nil
}

func GetUsersInGroup(svc *iam.IAM, group *iam.Group) ([]*iam.User, error) {
	resp, err := svc.GetGroup(&iam.GetGroupInput{GroupName: group.GroupName})

	if err != nil {
		return []*iam.User{}, err
	}

	return resp.Users, nil
}

func GetAuthorizedKeys(username string) (*bytes.Buffer, error) {

	sess, err := session.NewSession()

	if err != nil {
		return nil, err
	}

	svc := iam.New(sess, &aws.Config{Region: aws.String("us-east-1")})

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
