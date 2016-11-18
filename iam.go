package main

import (
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
