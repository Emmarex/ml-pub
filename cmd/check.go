/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:     "check",
	Aliases: []string{"validate"},
	Short:   "Validate mlpub Yaml file",
	Long:    `Validate mlpub Yaml file`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Printf("\x1b[31;1m%s\x1b[0m\n", "Config file does not exist")
			os.Exit(1)
		}

		yamlFile, err := ioutil.ReadFile(configFile)
		CheckIfError(err)
		data := PubConfiguration{}

		err = yaml.Unmarshal(yamlFile, &data)
		CheckIfError(err)

		fmt.Println(data)
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
	checkCmd.Flags().StringP("config", "c", "./mlpub.yaml", "Config file path")
}
