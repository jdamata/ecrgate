package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var (
	undefined int
	info      int
	low       int
	medium    int
	high      int
	critical  int
	rootCmd   = &cobra.Command{
		Use:   "ecrgate",
		Short: "Build, push and gate docker image promotion to ECR",
		Run:   main,
	}
)

func bindFlags(flags []string) {
	for _, value := range flags {
		viper.BindPFlag(value, rootCmd.Flags().Lookup(value))
	}
}

// Execute executes the root command.
func Execute(version string) error {
	rootCmd.Version = version
	rootCmd.Flags().StringP("dockerfile", "d", "./Dockerfile", "Path to Dockerfile")
	rootCmd.Flags().StringP("image", "i", "", "Existing docker image to pull down")
	rootCmd.Flags().StringP("repo", "r", "", "ECR repo to create and push image to")
	rootCmd.Flags().StringP("tag", "t", "latest", "Docker tag to build")
	rootCmd.Flags().StringSliceP("build_args", "b", []string{}, "List of docker build args")
	rootCmd.Flags().BoolP("clean", "c", false, "Delete image from ECR if scan fails threshold")
	rootCmd.Flags().BoolP("disable_scan", "s", false, "Disable scanning of image")
	rootCmd.Flags().StringSliceP("accounts", "a", []string{}, "List of AWS account ids to allow pulling images from")
	rootCmd.Flags().IntVar(&undefined, "undefined", 100, "Acceptable threshold for UNDEFINED level results")
	rootCmd.Flags().IntVar(&info, "info", 25, "Acceptable threshold for INFORMATIONAL level results")
	rootCmd.Flags().IntVar(&low, "low", 10, "Acceptable threshold for LOW level results")
	rootCmd.Flags().IntVar(&medium, "medium", 5, "Acceptable threshold for MEDIUM level results")
	rootCmd.Flags().IntVar(&high, "high", 3, "Acceptable threshold for HIGH level results")
	rootCmd.Flags().IntVar(&critical, "critical", 1, "Acceptable threshold for CRITICAL level results")
	rootCmd.MarkFlagRequired("repo")
	bindFlags([]string{"dockerfile", "image", "tag", "repo", "clean", "disable_scan", "accounts", "info", "low", "medium", "high", "critical", "undefined", "build_args"})
	return rootCmd.Execute()
}

func main(cmd *cobra.Command, args []string) {
	// Logging config
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	// Client initialization
	ctx := context.Background()
	svc := ecr.New(session.New())
	docker, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create docker client. Is docker running? - %s", err)
	}

	// Get viper configs
	repo := viper.GetString("repo")
	tag := viper.GetString("tag")

	if viper.GetString("image") != "" {
		// If image is defined, lets set the tag to the same as the image we are pulling
		image := strings.Split(viper.GetString("image"), ":")
		// If image has tag defined, use that tag. Otherwise it will default to latest
		if len(image) > 1 {
			tag = image[1]
		}
	}

	// Create ECR repo and add ecr policy
	createRepo(svc, repo)
	accountIDs := viper.GetStringSlice("accounts")
	if len(accountIDs) >= 1 {
		allowRepoPull(svc, accountIDs, repo)
	}

	// Authenticate to ECR repo, docker build and docker push
	ecrToken, imageURL := ecrCreds(svc, repo+":"+tag)

	// Build OR pull down image
	if viper.GetString("image") != "" {
		image := viper.GetString("image")
		log.Info("image flag passed, going to pull down image instead of building local Dockerfile")
		dockerPull(ctx, docker, image)
		dockerTag(ctx, docker, image, imageURL)
	} else {
		dockerBuild(ctx, docker, imageURL)
	}

	// Push docker image to ECR
	dockerPush(ctx, docker, svc, ecrToken, imageURL)

	if !viper.GetBool("disable_scan") {
		results := getScanResults(svc, repo, tag)
		resultLink := fmt.Sprintf("https://console.aws.amazon.com/ecr/repositories/%v/%v/%v/image/%v/scan-results/?region=%v",
			"private", // Need to support both public and private repos
			*results.RegistryId,
			*results.RepositoryName,
			*results.ImageId.ImageDigest,
			os.Getenv("AWS_REGION"))

		// Poll and pull ECR scan results
		allowedThresholds := map[string]int64{
			ecr.FindingSeverityUndefined:     viper.GetInt64("undefined"),
			ecr.FindingSeverityInformational: viper.GetInt64("info"),
			ecr.FindingSeverityLow:           viper.GetInt64("low"),
			ecr.FindingSeverityMedium:        viper.GetInt64("medium"),
			ecr.FindingSeverityHigh:          viper.GetInt64("high"),
			ecr.FindingSeverityCritical:      viper.GetInt64("critical"),
		}

		// Compare scan results to specified thresholds
		failedScan, failedLevels := compareThresholds(allowedThresholds, results.ImageScanFindings.FindingSeverityCounts)
		if failedScan {
			log.Errorf("Scan failed due to exceeding threshold levels: %v", failedLevels)

			// Delete docker image if scan threshold exceeded AND clean flag specified
			if viper.GetBool("clean") {
				log.Info("Clean specified. Deleting image from ecr")
				deleteImage(svc, repo, tag)
			}
			// Purposely return an error code to fail CI builds
			log.Infof("Scan result link: %v", resultLink)
			os.Exit(1)
		} else {
			log.Info("Scan passed!")
			log.Infof("Scan result link: %v", resultLink)
		}
	} else {
		log.Info("Skipping scan")
	}
}
