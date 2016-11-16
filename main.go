package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)

func get_instances() {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := ec2.New(sess, &aws.Config{Region: aws.String("us-east-1")})

	resp, err := svc.DescribeInstances(nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("> Number of reservation sets: ", len(resp.Reservations))

	for idx, res := range resp.Reservations {
		fmt.Println("  > Number of instances: ", len(res.Instances))
		for _, inst := range resp.Reservations[idx].Instances {
			fmt.Println("    - Instance ID: ", *inst.InstanceId)
		}
	}
}

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

func main() {
	get_groups()
}
