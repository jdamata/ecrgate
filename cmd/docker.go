package cmd

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/service/ecr"
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"
)

//writes from the build response to the log
func dockerLogOutput(reader io.ReadCloser) {
	defer reader.Close()
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	err := jsonmessage.DisplayJSONMessagesStream(reader, os.Stderr, termFd, isTerm, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

// DockerBuild - builds and tags a docker image
func dockerBuild(ctx context.Context, docker *client.Client, dockerfile string, image string) {
	// Docker build config
	buildOpts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{image},
	}
	// Create docker context
	buildCtx, err := archive.TarWithOptions(dockerfile, &archive.TarOptions{})
	if err != nil {
		log.Fatalf("Failed to build docker context - %s", err)
	}
	log.Info("Building docker image")
	out, err := docker.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		log.Fatalf("Failed to build docker image - %s", err)
	}
	dockerLogOutput(out.Body)
}

// DockerPush - pushes an image to the ECR repo
func dockerPush(ctx context.Context, docker *client.Client, svc *ecr.ECR, ecrToken string, imageDest string) {
	// Push docker image
	log.Info("Pushing image: ", imageDest)
	out, err := docker.ImagePush(ctx, imageDest, types.ImagePushOptions{RegistryAuth: ecrToken})
	if err != nil {
		log.Fatalf("Failed to initalize push image to docker daemon: %v. Error: %s", imageDest, err)
	}
	dockerLogOutput(out)
}
