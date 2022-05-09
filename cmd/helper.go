
package cmd

var cloudServiceProviders = []string{"AWS", "Azure", "Google Cloud"}

type PubConfiguration struct {
	Name				string
	ModelPath 			string `yaml:"model_path"`
	PreProcessor		string `yaml:"pre_processor"`
	CloudService		string `yaml:"cloud_service"`
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