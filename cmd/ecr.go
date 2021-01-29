package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type policyDocument struct {
	Version   string
	Statement []statementEntry
}

type statementEntry struct {
	Effect    string
	Action    []string
	Principal principalEntry
	Sid       string
}

type principalEntry struct {
	AWS []string
}

// allowRepoPull: Adds a ECR policy to allow a different AWS account to pull images
func allowRepoPull(svc *ecr.ECR, accountIDs []string, ecrRepo string) {
	policy := policyDocument{
		Version: "2008-10-17",
		Statement: []statementEntry{
			{
				Sid:    "allow-pull-from-aws-accounts",
				Effect: "Allow",
				Action: []string{
					"ecr:BatchCheckLayerAvailability",
					"ecr:BatchGetImage",
					"ecr:GetDownloadUrlForLayer",
				},
				Principal: principalEntry{
					AWS: []string{},
				},
			},
		},
	}
	for _, v := range accountIDs {
		accountID := fmt.Sprintf("arn:aws:iam::%s:root", v)
		policy.Statement[0].Principal.AWS = append(policy.Statement[0].Principal.AWS, accountID)
	}

	jsonPolicy, _ := json.Marshal(policy)
	jsonPolicyString := string(jsonPolicy)

	policyInput := ecr.SetRepositoryPolicyInput{
		PolicyText:     &jsonPolicyString,
		RepositoryName: &ecrRepo,
	}

	out, err := svc.SetRepositoryPolicy(&policyInput)
	if err != nil {
		log.Errorf("Failed to add ECR policy - %s", err)
	} else {
		log.Infof("Added ECR policy - %s", out.String())
	}
}

// CreateRepo - Creates an ECR repo if one does not exist
func createRepo(svc *ecr.ECR, ecrRepo string) {
	// Repo config
	enableScan := true
	enableScanning := ecr.ImageScanningConfiguration{
		ScanOnPush: &enableScan,
	}
	tagMutability := "MUTABLE"
	repoConfig := &ecr.CreateRepositoryInput{
		RepositoryName:             aws.String(ecrRepo),
		ImageScanningConfiguration: &enableScanning,
		ImageTagMutability:         &tagMutability,
	}

	// Check if repo exists
	existingRepos, err := svc.DescribeRepositories(&ecr.DescribeRepositoriesInput{})
	if err != nil {
		log.Fatalf("Cannot get list of existing repositories - %v", err)
	}
	for _, repo := range existingRepos.Repositories {
		if ecrRepo == *repo.RepositoryName {
			log.Info("ECR repo already exists. Skipping creation")
			if !*repo.ImageScanningConfiguration.ScanOnPush {
				log.Warning("ECR repo has scanning configuration disabled. Enabling this now...")
				scanConfig := ecr.PutImageScanningConfigurationInput{
					ImageScanningConfiguration: &enableScanning,
					RepositoryName:             repo.RepositoryName,
				}
				_, err := svc.PutImageScanningConfiguration(&scanConfig)
				if err != nil {
					log.Fatalf("Failed to enable scanning configuration - %v", err)
				} else {
					log.Info("Enabled scanning configuration.")
				}
			}
			return
		}
	}

	// Create repo
	result, err := svc.CreateRepository(repoConfig)
	if err != nil {
		log.Fatalf("Failed to create ECR repo - %s", err)
	} else {
		log.Infof("Created ECR repository - %s", result.String())
	}
}

// ecrCreds: Generate docker AuthStr for authentication
func ecrCreds(svc *ecr.ECR, image string) (string, string) {
	// Get ECR docker credential
	tokenOutput, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		log.Fatalf("Failed to obtain docker credentials - %s", err)
	}
	token := tokenOutput.AuthorizationData[0].AuthorizationToken
	imageURL, _ := url.Parse(*tokenOutput.AuthorizationData[0].ProxyEndpoint)
	imageDest := imageURL.Host + "/" + image

	// Convert token into proper authStr for docker login
	authInfoBytes, _ := base64.StdEncoding.DecodeString(*token)
	authInfo := strings.Split(string(authInfoBytes), ":")
	auth := struct {
		Username string
		Password string
	}{
		Username: authInfo[0],
		Password: authInfo[1],
	}
	authBytes, _ := json.Marshal(auth)
	authStr := base64.StdEncoding.EncodeToString(authBytes)
	return authStr, imageDest
}

// getScanResults: Grabs scan results
func getScanResults(svc *ecr.ECR, repo string, imageTag string) *ecr.DescribeImageScanFindingsOutput {
	// Constuct scan findings config
	imageID := ecr.ImageIdentifier{
		ImageTag: &imageTag,
	}
	scanConfig := ecr.DescribeImageScanFindingsInput{
		ImageId:        &imageID,
		RepositoryName: &repo,
	}

	// Poll for scan results
	for {
		out, err := svc.DescribeImageScanFindings(&scanConfig)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				switch awsErr.Code() {
				// if the image does not exist in the repo, we should retry after 2 seconds.
				case ecr.ErrCodeImageNotFoundException:
					log.Errorf("%s - Retrying in 2 seconds", err)
					sleepHelper(2)
				}
			}
		}
		if *out.ImageScanStatus.Status == string("IN_PROGRESS") {
			log.Info("Scan IN_PROGRESS")
			sleepHelper(5)
		} else if *out.ImageScanStatus.Status == string("FAILED") {
			log.Fatalf("ECR scan failed - %s", *out.ImageScanStatus.Description)
		} else if *out.ImageScanStatus.Status == string("COMPLETE") {
			return out
		}
	}
}

func sleepHelper(sleep int) {
	log.Infof("Sleeping for %v seconds", sleep)
	time.Sleep(time.Duration(sleep * 1000000000))
}

// deleteImage: Deletes image from ECR repo
func deleteImage(svc *ecr.ECR, repo string, imageTag string) {
	deleteConfig := &ecr.BatchDeleteImageInput{
		ImageIds: []*ecr.ImageIdentifier{
			{
				ImageTag: aws.String(imageTag),
			},
		},
		RepositoryName: aws.String(repo),
	}
	_, err := svc.BatchDeleteImage(deleteConfig)
	if err != nil {
		log.Fatalf("Failed to delete image: %v, %s", imageTag, err.Error())
	} else {
		log.Info("Cleaned image from repo.")
	}
}
