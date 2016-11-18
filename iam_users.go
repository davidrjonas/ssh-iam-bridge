package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

func iamUserNames(users []*iam.User) []string {
	names := make([]string, len(users))
	for idx, u := range users {
		names[idx] = aws.StringValue(u.UserName)
	}
	return names
}

func GetUser(username string) (*iam.User, error) {

	svc := getIamService()

	resp, err := svc.GetUser(GetUserInput{UserName: username})
	if err != nil {
		return nil, err
	}

	return resp.User, nil
}
