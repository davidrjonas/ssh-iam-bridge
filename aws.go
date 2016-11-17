package main

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

var iam_svc *iam.IAM

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

func getIamGroups(prefixes []string) ([]*iam.Group, error) {

	svc := getIamService()

	resp, err := svc.ListGroups(nil)

	if err != nil {
		return []*iam.Group{}, err
	}

	return filterGroups(resp.Groups, func(g *iam.Group) bool {
		return isPrefixedBy(g.GroupName, prefixes)
	}), nil
}

func getIamGroupUsers(group *iam.Group) ([]*iam.User, error) {

	svc := getIamService()

	resp, err := svc.GetGroup(&iam.GetGroupInput{GroupName: group.GroupName})

	if err != nil {
		return []*iam.User{}, err
	}

	return resp.Users, nil
}

func iamUserNames(users []*iam.User) []string {
	names := make([]string, len(users))
	for idx, u := range users {
		names[idx] = aws.StringValue(u.UserName)
	}
	return names
}

func onlyActiveKeys(keys []*iam.SSHPublicKeyMetadata) []*iam.SSHPublicKeyMetadata {

	var active []*iam.SSHPublicKeyMetadata

	for _, key := range keys {
		if aws.StringValue(key.Status) != iam.StatusTypeActive {
			continue
		}
		active = append(active, key)
	}

	return active
}

func getActiveSshPublicKeys(username string) ([]*iam.SSHPublicKeyMetadata, error) {

	svc := getIamService()

	resp, err := svc.ListSSHPublicKeys(&iam.ListSSHPublicKeysInput{UserName: &username})

	if err != nil {
		return nil, err
	}

	return onlyActiveKeys(resp.SSHPublicKeys), nil
}

func getSshEncodePublicKey(username, key_id *string) (*string, error) {

	svc := getIamService()

	keyref, err := svc.GetSSHPublicKey(&iam.GetSSHPublicKeyInput{
		SSHPublicKeyId: key_id,
		UserName:       username,
		Encoding:       aws.String(iam.EncodingTypeSsh),
	})

	if err != nil {
		return nil, err
	}

	return keyref.SSHPublicKey.SSHPublicKeyBody, nil
}
