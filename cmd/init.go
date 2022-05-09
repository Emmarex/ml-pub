/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"os"


	"github.com/spf13/cobra"

	"github.com/manifoldco/promptui"

	"gopkg.in/yaml.v3"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path]",
	Aliases: []string{"initialize", "initialise", "create"},
	Short: "Initialize a mlpub project",
	Long: `Initialize a mlpub project`,
	Run: func(cmd *cobra.Command, args []string) {
		var projectPath, projectName, modelPath, preProcessor, cloudService string
		projectPath, err := os.Getwd()
		
		if len(args) > 0 {
			if args[0] != "." {
				projectPath = args[0]
			}
		}

		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			// create directory
			if err := os.Mkdir(projectPath, 0754); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if projectName == "" {
			prompt := promptui.Prompt{
				Label: "Project Name",
			}
			result, err := prompt.Run()
			if err != nil {
				fmt.Println("You have to enter a project name.")
				return
			}
			projectName = result
		}

		if modelPath == "" {
			prompt := promptui.Prompt{
				Label:   "Model Path",
				Validate: func(input string) error {
					if _, err := os.Stat(input); os.IsNotExist(err) {
						return errors.New("Invalid Path")
					}
					return nil
				},
			}
			result, err := prompt.Run()
			if err != nil {
				fmt.Println("You have to enter your model file path.")
				return
			}
			modelPath = result
		}

		if cloudService == "" {
			prompt := promptui.Select{
				Label: "Cloud service",
				Items: cloudServiceProviders,
			}
			_, result, err := prompt.Run()
			if err != nil {
				fmt.Println("You have to select a Cloud service to deploy your project")
				return
			}
			cloudService = result
		}

		if preProcessor == "" {
			preProcessor = fmt.Sprintf("%s/%s", projectPath, "pre_processor.py")
			fmt.Println(preProcessor)
			os.WriteFile(preProcessor, preProcessorTemplate(), 0754)
		}

		data := PubConfiguration{projectName,modelPath,preProcessor,cloudService}
		configByte, err := yaml.Marshal(&data)
        if err != nil {
			fmt.Println(err)
			os.Exit(1)
        }
		os.WriteFile(fmt.Sprintf("%s/%s", projectPath, "mlpub.yaml") , configByte, 0754)
		fmt.Println(fmt.Sprintf("Project Initialized successfully at %s", projectPath))
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
	initCmd.Flags().StringP("pre_processor", "p", "", "Python preprocessor file")
	initCmd.Flags().StringP("cloud_service", "c", "", "Cloud service to host project")
}
