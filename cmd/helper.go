package cmd

import (
	"fmt"
	"os"

	"strings"
)

var cloudServiceProviders = []string{"AWS", "Azure", "Google Cloud"}

type PubConfiguration struct {
	Name         string
	ModelPath    string `yaml:"model_path"`
	PreProcessor string `yaml:"pre_processor"`
	CloudService string `yaml:"cloud_service"`
}

func preProcessorTemplate() []byte {
	return []byte(`
"""
This is a sample python pre processing file for your model
"""
from typing import Tuple, Any
	
def main(data) -> Tuple[bool, Any]:
	"""
	do not change the name of this method
	"""
	# return the pre-processed version of your data here
	return True, data
	`)
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
