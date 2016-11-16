package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

func get_groups() {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := iam.New(sess, &aws.Config{Region: aws.String("us-east-1")})

	resp, err := svc.ListGroups(nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("> Number of groups: ", len(resp.Groups))

	for _, res := range resp.Groups {
		//fmt.Println("  > Group name: ", res.String())
		fmt.Println("  > Group name: ", *res.GroupName)
		fmt.Println("          path: ", *res.Path)
		fmt.Println("          id:   ", *res.GroupId)
		fmt.Println("          arn:  ", *res.Arn)

		resp, err := svc.GetGroup(&iam.GetGroupInput{GroupName: res.GroupName})
		if err != nil {
			panic(err)
		}

		fmt.Println("          usrs#:", len(resp.Users))

		fmt.Printf("          usrs: ")
		for _, userref := range resp.Users {
			fmt.Printf("%s ", *userref.UserName)
		}

		fmt.Println("")
	}
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

		fmt.Fprintln(&out, "# Key id: ", *keyref.SSHPublicKeyId)
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
