package directory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ARN struct {
	Partition string
	Service   string
	Region    string
	Account   string
	Resource  string
}

type IamInfo struct {
	Code               string `json:"Code"`
	LastUpdatee        string `json:"LastUpdated"`
	InstanceProfileArn string `json:"InstanceProfileArn"`
	InstanceProfileId  string `json:"InstanceProfileId"`
}

func parseArn(arn string) ARN {
	//  "arn:aws:iam::756413706286:instance-profile/bastion",
	parts := strings.Split(arn, ":")

	return ARN{
		Partition: parts[1],
		Service:   parts[2],
		Region:    parts[3],
		Account:   parts[4],
		Resource:  parts[5],
	}
}

// curl http://169.254.169.254/latest/meta-data/iam/info
// {
//  "Code" : "Success",
//  "LastUpdated" : "2016-11-16T18:52:29Z",
//  "InstanceProfileArn" : "arn:aws:iam::756413706286:instance-profile/bastion",
//  "InstanceProfileId" : "AIPAIHHSYB75V3MLIVTG6"
//}
func GetRole() (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/iam/info")
	if err != nil {
		return "", err
	}

	// If there is no IAM role then info will return 404
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	var info IamInfo

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}

	arn := parseArn(info.InstanceProfileArn)
	parts := strings.SplitN(arn.Resource, "/", 2)

	if len(parts) != 2 {
		return "", fmt.Errorf("arn resource in unknown format; Resource=%s", arn.Resource)
	}

	return parts[1], nil
}
