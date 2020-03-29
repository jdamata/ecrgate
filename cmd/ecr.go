package cmd

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
)

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
func getScanResults(svc *ecr.ECR, repo string, imageTag string) map[string]*int64 {
	// Constuct scan findings config
	imageID := ecr.ImageIdentifier{
		ImageTag: &imageTag,
	}
	scanConfig := ecr.DescribeImageScanFindingsInput{
		ImageId:        &imageID,
		RepositoryName: &repo,
	}
	// Sleep before checking for results. Probably a better way to do this
	sleep := 5
	log.Infof("Sleeping for %v seconds", sleep)
	time.Sleep(time.Duration(sleep * 1000000000))
	// Grab scan results
	out, err := svc.DescribeImageScanFindings(&scanConfig)
	if err != nil {
		log.Fatalf("Failed to get scan results - %s", err)
	}
	if *out.ImageScanStatus.Status == string("FAILED") {
		log.Fatalf("ECR scan failed - %s", *out.ImageScanStatus.Description)
	}
	return out.ImageScanFindings.FindingSeverityCounts
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
		log.Fatalf("Failed to delete image: %v, %s", imageTag, err.Error)
	} else {
		log.Info("Cleaned image from repo.")
	}
}
