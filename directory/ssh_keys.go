package directory

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

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

func GetActiveSshPublicKeys(username string) ([]*iam.SSHPublicKeyMetadata, error) {

	svc := getIamService()

	resp, err := svc.ListSSHPublicKeys(&iam.ListSSHPublicKeysInput{UserName: &username})

	if err != nil {
		return nil, err
	}

	return onlyActiveKeys(resp.SSHPublicKeys), nil
}

func GetSshEncodedPublicKey(username, keyID *string) (*string, error) {

	svc := getIamService()

	keyref, err := svc.GetSSHPublicKey(&iam.GetSSHPublicKeyInput{
		SSHPublicKeyId: keyID,
		UserName:       username,
		Encoding:       aws.String(iam.EncodingTypeSsh),
	})

	if err != nil {
		return nil, err
	}

	return keyref.SSHPublicKey.SSHPublicKeyBody, nil
}
