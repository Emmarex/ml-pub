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
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	cp "github.com/otiai10/copy"
	"gopkg.in/yaml.v3"
)

type AWSExtras struct {
	Region     string    `yaml:"region"`
	Layers     []*string `yaml:",flow"`
	S3Bucket   string    `yaml:"s3_bucket"`
	Runtime    string    `yaml:"runtime"`
	MemorySize int64     `yaml:"memory_size"`
	TimeOut    int64     `yaml:"time_out"`
	LambdaRole string    `yaml:"lambda_role"`
}

type PubConfiguration struct {
	Name         string
	ModelPath    string     `yaml:"model_path"`
	PreProcessor string     `yaml:"pre_processor"`
	CloudService string     `yaml:"cloud_service"`
	AWSExtras    *AWSExtras `yaml:"extras,omitempty"`
}

var cloudServiceProviders = []string{"AWS", "Google Cloud"}

var defaultAWSConfig = AWSExtras{
	Region:     "us-west-1",
	Layers:     []*string{},
	S3Bucket:   "",
	Runtime:    "python3.9",
	MemorySize: 256,
	TimeOut:    300,
	LambdaRole: "",
}

var defaultLambdaPolicy = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sts:AssumeRole"
            ],
            "Principal": {
                "Service": [
                    "lambda.amazonaws.com"
                ]
            }
        }
    ]
}`

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
	zipFilePath := randString(25) + ".zip"

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

func createConfigFile(pubConfig PubConfiguration, projectPath string) {
	configByte, err := yaml.Marshal(&pubConfig)
	CheckIfError(err)
	os.WriteFile(fmt.Sprintf("%s/%s", projectPath, "mlpub.yaml"), configByte, 0754)
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
	Info("Building application ... \n")
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

func createAWSbucket(pubConfig PubConfiguration, zipFileName string) string {
	bucketName := fmt.Sprintf("%s-mlpub-cli-managed-bucket", strings.ToLower(pubConfig.Name))
	Info("Creating AWS Bucket (%s) ...", bucketName)
	awsSession := session.Must(session.NewSession())
	svc := s3.New(awsSession, &aws.Config{Region: aws.String(pubConfig.AWSExtras.Region)})
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := svc.CreateBucket(input)
	os.RemoveAll(zipFileName)
	CheckIfError(err)
	return bucketName
}

func uploadZipFile(pubConfig PubConfiguration, zipFileName string) {
	Info("Uploading zip file ... \n")
	awsSession := session.Must(session.NewSession(&aws.Config{Region: aws.String(pubConfig.AWSExtras.Region)}))
	uploader := s3manager.NewUploader(awsSession)
	zipFile, err := os.Open(zipFileName)
	CheckIfError(err)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(pubConfig.AWSExtras.S3Bucket),
		Key:    aws.String(zipFileName),
		Body:   zipFile,
	})
	os.RemoveAll(zipFileName)
	CheckIfError(err)
}

func createLambdaRole(pubConfig PubConfiguration) string {
	Info("Creating lambda role ... \n")
	awsSession := session.Must(session.NewSession())
	svc := iam.New(awsSession, &aws.Config{Region: aws.String(pubConfig.AWSExtras.Region)})
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(defaultLambdaPolicy),
		Path:                     aws.String("/"),
		RoleName:                 aws.String(fmt.Sprintf("%s-role-mlpub-cli-managed", strings.ToLower(pubConfig.Name))),
	}
	result, err := svc.CreateRole(input)
	CheckIfError(err)
	return string(*result.Role.Arn)
}

func createAWSLambdaFunction(zipFileName string, pubConfig PubConfiguration) {
	Info("Creating lambda function ... \n")
	awsSession := session.Must(session.NewSession())
	svc := lambda.New(awsSession, &aws.Config{Region: aws.String(pubConfig.AWSExtras.Region)})
	input := &lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			S3Bucket: aws.String(pubConfig.AWSExtras.S3Bucket),
			S3Key:    aws.String(zipFileName),
		},
		Description:  aws.String("Lambda function created by mlpub"),
		FunctionName: aws.String(fmt.Sprintf("%s-function-%s", strings.ToLower(pubConfig.Name), randString(10))),
		Handler:      aws.String("app.handler"),
		MemorySize:   aws.Int64(pubConfig.AWSExtras.MemorySize),
		Layers:       pubConfig.AWSExtras.Layers,
		Publish:      aws.Bool(true),
		Role:         aws.String(pubConfig.AWSExtras.LambdaRole),
		Runtime:      aws.String(pubConfig.AWSExtras.Runtime),
		Timeout:      aws.Int64(pubConfig.AWSExtras.TimeOut),
	}

	_, err := svc.CreateFunction(input)
	CheckIfError(err)
}
