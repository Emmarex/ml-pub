/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/manifoldco/promptui"

	"github.com/go-git/go-git/v5"

	cp "github.com/otiai10/copy"

	"golang.org/x/exp/slices"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init [path]",
	Aliases: []string{"initialize", "initialise", "create"},
	Short:   "Initialize a mlpub project",
	Long:    `Initialize a mlpub project`,
	Run: func(cmd *cobra.Command, args []string) {
		// get flag values
		projectName, _ := cmd.Flags().GetString("project_name")
		modelPath, _ := cmd.Flags().GetString("model_path")
		preProcessor, _ := cmd.Flags().GetString("pre_processor")
		cloudService, _ := cmd.Flags().GetString("cloud_service")

		projectPath, err := os.Getwd()

		CheckIfError(err)

		if len(args) > 0 {
			if args[0] != "." {
				CheckArgs("<c>")
				projectPath = filepath.Clean(args[0])
			}
		}

		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			// create directory
			if err := os.Mkdir(projectPath, 0754); err != nil {
				fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if projectName == "" {
			prompt := promptui.Prompt{
				Label: "Project Name",
			}
			result, err := prompt.Run()
			CheckIfError(err)
			projectName = result
		}

		if modelPath == "" {
			prompt := promptui.Prompt{
				Label: "Model Path",
				Validate: func(input string) error {
					if _, err := os.Stat(input); os.IsNotExist(err) {
						return errors.New("invalid model path")
					}
					return nil
				},
			}
			result, err := prompt.Run()
			CheckIfError(err)
			modelPath = result
		} else {
			_, err := os.Stat(modelPath)
			CheckIfError(err)
		}

		if cloudService == "" {
			prompt := promptui.Select{
				Label: "Cloud service",
				Items: cloudServiceProviders,
			}
			_, result, err := prompt.Run()
			CheckIfError(err)
			cloudService = result
		} else {
			if !slices.Contains(cloudServiceProviders, cloudService) {
				Warning("%s is not a valid Cloud service provider option", cloudService)
			}
		}

		if preProcessor != "" {
			_, err := os.Stat(preProcessor)
			CheckIfError(err)
		}

		gitUrl := fmt.Sprintf("https://github.com/Emmarex/mlpub-template-%s", strings.ToLower(cloudService))

		// Clone the given repository to the given directory
		Info("cloning %s (process may take a moment) ...", gitUrl)

		_, err = git.PlainClone(projectPath, false, &git.CloneOptions{
			URL:               gitUrl,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		CheckIfError(err)

		os.RemoveAll(fmt.Sprintf("%s/%s", projectPath, ".git"))

		os.RemoveAll(fmt.Sprintf("%s/%s", projectPath, "LICENSE"))

		if preProcessor != "" {
			os.RemoveAll(fmt.Sprintf("%s/%s", projectPath, "pre_processor.py"))
			err = cp.Copy(preProcessor, fmt.Sprintf("%s/%s", projectPath, "pre_processor.py"))
			CheckIfError(err)
		} else {
			preProcessor = "pre_processor.py"
		}

		modelFileName := fmt.Sprintf("model%s", filepath.Ext(modelPath))
		err = cp.Copy(modelPath, filepath.Join(projectPath, modelFileName))
		CheckIfError(err)
		modelPath = modelFileName

		data := PubConfiguration{projectName, modelPath, preProcessor, cloudService, new(AWSExtras)}

		if cloudService == "AWS" {
			data.AWSExtras = &defaultAWSConfig
		}

		createConfigFile(data, projectPath)

		Info("Project Initialized successfully at %s", projectPath)

		Info("\n\n\t cd %s", projectPath)
		Info("\n\t mlpub deploy\n")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	initCmd.Flags().StringP("project_name", "n", "", "Project name")
	initCmd.Flags().StringP("model_path", "m", "", "Model Path")
	initCmd.Flags().StringP("pre_processor", "p", "", "Python pre-processor file path")
	initCmd.Flags().StringP("cloud_service", "c", "", "Cloud service to host project")
}
