package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

var cloudServiceProviders = []string{"AWS", "Google Cloud"}

type PubConfiguration struct {
	Name         string
	ModelPath    string `yaml:"model_path"`
	PreProcessor string `yaml:"pre_processor"`
	CloudService string `yaml:"cloud_service"`
}

func CheckArgs(arg ...string) {
	if len(os.Args) < len(arg)+1 {
		Warning("Usage: %s %s", os.Args[0], strings.Join(arg, " "))
		os.Exit(1)
	}
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func dockerize(projectPath string, projectName string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*360)
	defer cancel()

	tarFile, err := archive.TarWithOptions(projectPath, &archive.TarOptions{})
	CheckIfError(err)

	opts := types.ImageBuildOptions{
		Dockerfile:     "Dockerfile",
		Tags:           []string{projectName},
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		NoCache:        false,
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	CheckIfError(err)

	buildResponse, err := dockerClient.ImageBuild(ctx, tarFile, opts)
	CheckIfError(err)

	defer buildResponse.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(buildResponse.Body, os.Stderr, termFd, isTerm, nil)
}

func pushImageToECR(projectName string, awsRegion string) {
	awsSession := session.Must(session.NewSession())
	ecrClient := ecr.New(awsSession, &aws.Config{Region: aws.String(awsRegion)})
	result, err := ecrClient.PutImage(&ecr.PutImageInput{
		RepositoryName: aws.String(projectName),
		ImageManifest:  aws.String(""),
	})
	CheckIfError(err)
	fmt.Println(result)
}
