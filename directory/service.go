package directory

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

var iamSvc *iam.IAM

func getIamService() *iam.IAM {
	if iamSvc != nil {
		return iamSvc
	}

	sess, err := session.NewSession()

	if err != nil {
		panic(err)
	}

	iamSvc := iam.New(sess, &aws.Config{Region: aws.String("us-east-1")})

	return iamSvc
}
