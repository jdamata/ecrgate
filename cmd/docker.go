package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecr"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"
)

//dockerLogOutput: Formats docker output and displays to terminal
func dockerLogOutput(reader io.ReadCloser) {
	defer reader.Close()
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	err := jsonmessage.DisplayJSONMessagesStream(reader, os.Stderr, termFd, isTerm, nil)

	if err != nil {
		log.Fatalf(err.Error())
	}
}

func parseDockerArgs(buildArgsInput []string) map[string]*string {
	var buildArgs = map[string]*string{}

	for _, v := range buildArgsInput {
		// split each args on equal sign
		arg := strings.Split(v, "=")

		if len(arg) == 1 && arg[0] == "s" {
			log.Fatalf("Failed to parse build args. Build args must be in format foo=bar,baz=qux")
		}

		buildArgs[arg[0]] = &arg[1]
	}

	return buildArgs
}

// dockerBuild: builds and tags a docker image
func dockerBuild(ctx context.Context, docker *client.Client, image string) {
	dockerfile := viper.GetString("dockerfile")

	// Docker build config
	buildOpts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{image},
	}

	if buildArg := viper.GetStringSlice("build_args"); len(buildArg) >= 1 {
		buildOpts.BuildArgs = parseDockerArgs(buildArg)
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

// dockerPush: pushes an image to the ECR repo
func dockerPush(ctx context.Context, docker *client.Client, svc *ecr.ECR, ecrToken string, imageDest string) {
	// Push docker image
	log.Info("Pushing image: ", imageDest)
	out, err := docker.ImagePush(ctx, imageDest, types.ImagePushOptions{RegistryAuth: ecrToken})
	if err != nil {
		log.Fatalf("Failed to initialize push image to docker daemon: %v. Error: %s", imageDest, err)
	}
	dockerLogOutput(out)
}

// dockerPull: Pull docker iamge down
func dockerPull(ctx context.Context, docker *client.Client, image string) {
	log.Info("Pulling image: ", image)
	out, err := docker.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		log.Fatalf("Failed to pull docker image: %v. Error: %s", image, err)
	}
	dockerLogOutput(out)
}

// dockerTag: Pull docker iamge down
func dockerTag(ctx context.Context, docker *client.Client, image string, imageDest string) {
	log.Infof("Tagging docker image: %v with tag: %v", image, imageDest)
	err := docker.ImageTag(ctx, image, imageDest)
	if err != nil {
		log.Fatalf("Failed to tag docker image: %s", err)
	}
}
