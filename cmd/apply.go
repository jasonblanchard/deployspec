/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/jasonblanchard/deployspec/deployspec-sdk"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filepath, err := cmd.Flags().GetString(("file"))

		if err != nil {
			return err
		}

		yamlFile, err := ioutil.ReadFile(filepath)

		if err != nil {
			return err
		}

		spec := &deployspec.DeploySpec{}
		err = yaml.Unmarshal(yamlFile, spec)

		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Fatal(err)
		}

		lambdaclient := lambda.NewFromConfig(cfg)

		reconciler := &deployspec.Reconciler{
			Client: lambdaclient,
		}

		finalAppSpec, err := reconciler.Reconcile(spec)

		finalAppSpecString, err := yaml.Marshal(finalAppSpec)

		if err != nil {
			return err
		}

		fmt.Println(string(finalAppSpecString))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// applyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// applyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	applyCmd.Flags().StringP("file", "f", "", "deployspec file")

}
