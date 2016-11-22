package directory

import (
	"fmt"
	"os"

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
		fmt.Fprintf(os.Stderr, "Failed to start a new AWS session; %s", err)
		os.Exit(1)
	}

	iamSvc := iam.New(sess, &aws.Config{Region: aws.String("us-east-1")})

	return iamSvc
}
