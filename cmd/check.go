/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:     "check [path]",
	Aliases: []string{"validate"},
	Short:   "Validate mlpub Yaml file",
	Long:    `Validate mlpub Yaml file`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, err := os.Getwd()

		CheckIfError(err)

		if len(args) > 0 {
			if args[0] != "." {
				CheckArgs("<c>")
				projectPath = filepath.Clean(args[0])
			}
		}

		configFile := filepath.Join(projectPath, "mlpub.yaml")

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			CheckIfError(err)
		}

		yamlFile, err := ioutil.ReadFile(configFile)
		CheckIfError(err)
		data := PubConfiguration{}

		err = yaml.Unmarshal(yamlFile, &data)
		CheckIfError(err)

		Info("\nValid Config file")
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
