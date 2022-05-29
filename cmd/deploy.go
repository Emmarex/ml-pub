/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:     "deploy [path]",
	Aliases: []string{"push"},
	Short:   "Deploy your Machine learning model",
	Long:    `Deploy your Machine learning model`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, err := os.Getwd()

		CheckIfError(err)

		if len(args) > 0 {
			if args[0] != "." {
				CheckArgs("<c>")
				projectPath = filepath.Clean(args[0])
			}
		}

		deployDirectory := filepath.Clean(fmt.Sprintf("%s/%s", projectPath, ".mlpub"))
		if _, err := os.Stat(deployDirectory); os.IsNotExist(err) {
			// create directory
			err = os.Mkdir(deployDirectory, 0754)
			CheckIfError(err)
		}

		configFile := filepath.Join(projectPath, "mlpub.yaml")

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Printf("\x1b[31;1m%s\x1b[0m\n", "Config file does not exist")
			os.Exit(1)
		}

		yamlFile, err := ioutil.ReadFile(configFile)
		CheckIfError(err)
		pubConfig := PubConfiguration{}

		err = yaml.Unmarshal(yamlFile, &pubConfig)
		CheckIfError(err)

		Info("Building application ... \n")
		InstallProjectPackages(projectPath, deployDirectory)

		zipFileName := zipFiles(deployDirectory)

		if pubConfig.CloudService == "AWS" && pubConfig.AWSExtras.S3Bucket == "" {
			bucketName := createAWSbucket(fmt.Sprintf("%s-mlpub-bucket", pubConfig.Name), pubConfig.AWSExtras.Region)
			pubConfig.AWSExtras.S3Bucket = bucketName
		}

		Info("Uploading zip file ... \n")
		uploadZipFile(pubConfig, zipFileName)

		Info("Creating lambda function ... \n")
		createAWSLambdaFunction(zipFileName, pubConfig)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
