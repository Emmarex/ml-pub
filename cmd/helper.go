package cmd

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"archive/zip"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	cp "github.com/otiai10/copy"
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

func randString(length int) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomString := make([]byte, length)
	for item := range randomString {
		randomString[item] = charset[seededRand.Intn(len(charset))]
	}
	return string(randomString)
}

func zipFiles(deployPath string) string {
	zipFilePath := randString(50) + ".zip"

	newZipFile, err := os.Create(zipFilePath)
	CheckIfError(err)
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)

	filepath.Walk(deployPath, func(filePath string, info os.FileInfo, err error) error {
		CheckIfError(err)
		if info.IsDir() {
			return nil
		}
		relPath := strings.TrimPrefix(filePath, deployPath+"/")
		zipFile, err := zipWriter.Create(relPath)
		CheckIfError(err)
		fsFile, err := os.Open(filePath)
		CheckIfError(err)
		_, err = io.Copy(zipFile, fsFile)
		CheckIfError(err)
		return nil
	})
	zipWriter.Close()
	return zipFilePath
}

func CheckPip() string {
	pipPath, err := exec.LookPath("pip3")
	if err != nil {
		pipPath, err := exec.LookPath("pip")
		CheckIfError(err)
		return pipPath
	}
	return pipPath
}

func InstallProjectPackages(projectPath string, deployPath string) {
	err := cp.Copy(projectPath, deployPath, cp.Options{
		Skip: func(src string) (bool, error) {
			return strings.Contains(src, ".mlpub"), nil
		},
	})
	CheckIfError(err)
	pipPath := CheckPip()
	Info("Installing requirements ... \n")
	err = exec.Command(pipPath, "install", "-r", fmt.Sprintf("%s/requirements.txt", deployPath), "-t", deployPath).Run()
	CheckIfError(err)
}

func createAWSbucket(bucketName string, awsRegion string) {
	awsSession := session.Must(session.NewSession())
	svc := s3.New(awsSession, &aws.Config{Region: aws.String(awsRegion)})
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := svc.CreateBucket(input)
	CheckIfError(err)
}

func uploadZipFile(bucketName string, awsRegion string, zipFileName string) {
	awsSession := session.Must(session.NewSession())
	svc := s3.New(awsSession, &aws.Config{Region: aws.String(awsRegion)})
	input := &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(strings.NewReader(zipFileName)),
		Bucket: aws.String(bucketName),
		Key:    aws.String(zipFileName),
	}
	result, err := svc.PutObject(input)
	CheckIfError(err)
	fmt.Println(result)
}

func createAWSLambdaFunction(awsRegion string, bucketName string, zipFileName string, projectData PubConfiguration) {
	awsSession := session.Must(session.NewSession())
	svc := lambda.New(awsSession, &aws.Config{Region: aws.String(awsRegion)})
	input := &lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			S3Bucket: aws.String(bucketName),
			S3Key:    aws.String(zipFileName),
		},
		Description: aws.String("Lambda function created by mlpub"),
		// Environment: &lambda.Environment{
		// 	Variables: map[string]*string{
		// 		"BUCKET": aws.String("my-bucket-1xpuxmplzrlbh"),
		// 		"PREFIX": aws.String("inbound"),
		// 	},
		// },
		FunctionName: aws.String(fmt.Sprintf("%s-function", projectData.Name)),
		Handler:      aws.String("app.handler"),
		MemorySize:   aws.Int64(256),
		Publish:      aws.Bool(true),
		// Role:         aws.String("arn:aws:iam::123456789012:role/lambda-role"),
		Runtime: aws.String("nodejs12.x"),
		// Tags: map[string]*string{
		// 	"DEPARTMENT": aws.String("Assets"),
		// },
		Timeout: aws.Int64(15),
		TracingConfig: &lambda.TracingConfig{
			Mode: aws.String("Active"),
		},
	}

	result, err := svc.CreateFunction(input)
	CheckIfError(err)
	fmt.Println(result)
}
