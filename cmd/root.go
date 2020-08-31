package cmd

import (
	"os"

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
	rootCmd.Flags().StringP("dockerfile", "d", ".", "Path to Dockerfile")
	rootCmd.Flags().StringP("repo", "r", "", "ECR repo to create and push image to")
	rootCmd.Flags().StringP("tag", "t", "latest", "Docker tag to build")
	rootCmd.Flags().BoolP("clean", "c", false, "Delete image from ECR if scan fails threshold")
	rootCmd.Flags().BoolP("scan", "s", true, "Enable scanning of image")
	rootCmd.Flags().StringSliceP("accounts", "a", []string{}, "List of AWS account ids to allow pulling images from")
	rootCmd.Flags().IntVar(&undefined, "undefined", 100, "Acceptable threshold for UNDEFINED level results")
	rootCmd.Flags().IntVar(&info, "info", 25, "Acceptable threshold for INFORMATIONAL level results")
	rootCmd.Flags().IntVar(&low, "low", 10, "Acceptable threshold for LOW level results")
	rootCmd.Flags().IntVar(&medium, "medium", 5, "Acceptable threshold for MEDIUM level results")
	rootCmd.Flags().IntVar(&high, "high", 3, "Acceptable threshold for HIGH level results")
	rootCmd.Flags().IntVar(&critical, "critical", 1, "Acceptable threshold for CRITICAL level results")
	rootCmd.MarkFlagRequired("repo")
	bindFlags([]string{"dockerfile", "tag", "repo", "clean", "scan", "accounts", "info", "low", "medium", "high", "critical", "undefined"})
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
	image := repo + ":" + tag

	// Create ECR repo and add ecr policy
	createRepo(svc, repo)
	accountIDs := viper.GetStringSlice("accounts")
	if len(accountIDs) >= 1 {
		allowRepoPull(svc, accountIDs, repo)
	}

	// Authenticate to ECR repo, docker build and docker push
	ecrToken, imageURL := ecrCreds(svc, image)
	dockerBuild(ctx, docker, imageURL)
	dockerPush(ctx, docker, svc, ecrToken, imageURL)

	if viper.GetBool("scan") {
		// Poll and pull ECR scan results
		results := getScanResults(svc, repo, tag)
		allowedThresholds := map[string]int64{
			"UNDEFINED":     viper.GetInt64("undefined"),
			"INFORMATIONAL": viper.GetInt64("info"),
			"LOW":           viper.GetInt64("low"),
			"MEDIUM":        viper.GetInt64("medium"),
			"HIGH":          viper.GetInt64("high"),
			"CRITICAL":      viper.GetInt64("critical"),
		}

		// Compare scan results to specified thresholds
		failedScan, failedLevels := compareThresholds(allowedThresholds, results)
		if failedScan {
			log.Errorf("Scan failed due to exceeding threshold levels: %v", failedLevels)

			// Delete docker image if scan threshold exceeded AND clean flag specified
			if viper.GetBool("clean") {
				log.Info("Clean specified. Deleting image from ecr")
				deleteImage(svc, repo, tag)
				// Purposely return an error code to fail CI builds
				os.Exit(1)
			} else {
				os.Exit(1)
			}
		} else {
			log.Info("Scan passed!")
		}
	} else {
		log.Info("Skipping scan")
	}
}
