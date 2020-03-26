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
	rootCmd = &cobra.Command{
		Use:   "ecrgate",
		Short: "Build, push and gate docker image promotion to ECR",
		Run:   main,
	}
)

// Execute executes the root command.
func Execute(version string) error {
	rootCmd.Version = version
	rootCmd.Flags().StringP("dockerfile", "d", ".", "Path to Dockerfile")
	rootCmd.Flags().StringP("tag", "t", "latest", "Docker tag to build")
	rootCmd.Flags().StringP("repo", "r", "", "ECR repo to create and push image to")
	rootCmd.MarkFlagRequired("repo")
	viper.BindPFlag("dockerfile", rootCmd.Flags().Lookup("dockerfile"))
	viper.BindPFlag("tag", rootCmd.Flags().Lookup("tag"))
	viper.BindPFlag("repo", rootCmd.Flags().Lookup("repo"))
	return rootCmd.Execute()
}

func main(cmd *cobra.Command, args []string) {
	// Flags
	dockerfile := viper.GetString("dockerfile")
	tag := viper.GetString("tag")
	repo := viper.GetString("repo")
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
	createRepo(svc, repo)
	ecrToken, imageDest := ecrCreds(svc, repo+":"+tag)
	dockerBuild(ctx, docker, dockerfile, imageDest)
	dockerPush(ctx, docker, svc, ecrToken, imageDest)
	results := getScanResults(svc, tag, repo)
	log.Info(results)
}
