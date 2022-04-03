/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var fields []string
var inputFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "estable",
	Short: "estable: Get Elasticsearch as table view",
	Long:  `estable helps you get data from elasticsearch and show you desired fields in a terminal table view.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/estable.yaml)")
	rootCmd.PersistentFlags().StringArrayVarP(&fields, "field", "", []string{}, "Field to show in table")
	rootCmd.PersistentFlags().StringVarP(&inputFile, "file", "f", "", "File to use input instead of stdin")

	rootCmd.PersistentFlags().StringP("address", "a", "", "Elasticsearch address (http://localhost:9200)")
	viper.BindPFlag("address", rootCmd.PersistentFlags().Lookup("es_address"))
	rootCmd.PersistentFlags().StringP("user", "u", "", "Elasticsearch user")
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("es_user"))
	rootCmd.PersistentFlags().StringP("password", "p", "", "Elasticsearch password")
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("es_password"))
	rootCmd.PersistentFlags().StringP("index", "i", "", "Elasticsearch index")
	viper.BindPFlag("index", rootCmd.PersistentFlags().Lookup("es_index"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name "estable.yaml".
		viper.AddConfigPath(fmt.Sprint(home, "/.config"))
		viper.SetConfigName("estable")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
