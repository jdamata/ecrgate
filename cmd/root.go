package cmd

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var (
	info     int
	low      int
	medium   int
	high     int
	critical int
	rootCmd  = &cobra.Command{
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
	rootCmd.Flags().StringP("tag", "t", "latest", "Docker tag to build")
	rootCmd.Flags().StringP("repo", "r", "", "ECR repo to create and push image to")
	rootCmd.MarkFlagRequired("repo")
	rootCmd.Flags().IntVar(&info, "info", 50, "Acceptable threshold for INFORMATIONAL level results")
	rootCmd.Flags().IntVar(&low, "low", 20, "Acceptable threshold for LOW level results")
	rootCmd.Flags().IntVar(&medium, "medium", 10, "Acceptable threshold for MEDIUM level results")
	rootCmd.Flags().IntVar(&high, "high", 3, "Acceptable threshold for HIGH level results")
	rootCmd.Flags().IntVar(&critical, "critical", 0, "Acceptable threshold for CRITICAL level results")
	bindFlags([]string{"dockerfile", "tag", "repo", "info", "low", "medium", "high", "critical"})
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
	// Main
	tag := viper.GetString("tag")
	repo := viper.GetString("repo")
	createRepo(svc, repo)
	ecrToken, imageDest := ecrCreds(svc, repo+":"+tag)
	dockerBuild(ctx, docker, imageDest)
	dockerPush(ctx, docker, svc, ecrToken, imageDest)
	compareThresholds(getScanResults(svc, tag, repo))
}
