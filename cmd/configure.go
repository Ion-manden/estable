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
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure auth and hosts for elasticsearch",
	Long: `Set default host/address, user, password and index so you don't have to specify on every execution.

  Values are saves in yaml configfile by default located at ~/.config/estable.yaml, but can also be passed to commands.

  Values can alse be set as flags on commands.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("configure called")

		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Set elasticsearch address (http://localhost:9200):")
		fmt.Println("Current value:", viper.GetString("es_address"))
		text, err := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		viper.Set("es_address", text)

		fmt.Println("Set elasticsearch user:")
		fmt.Println("Current value:", viper.GetString("es_user"))
		text, err = reader.ReadString('\n')
		text = strings.TrimSpace(text)
		viper.Set("es_user", text)

		fmt.Println("Set elasticsearch password:")
		fmt.Println("Current value:", viper.GetString("es_password"))
		text, err = reader.ReadString('\n')
		text = strings.TrimSpace(text)
		viper.Set("es_password", text)
		// TODO: encode password - current issue is using is decoded after as default for flag
		// encoded := base64.StdEncoding.EncodeToString([]byte(text))
		// viper.Set("password", encoded)

		fmt.Println("Set elasticsearch default index:")
		fmt.Println("Current value:", viper.GetString("es_index"))
		text, err = reader.ReadString('\n')
		text = strings.TrimSpace(text)
		viper.Set("es_index", text)

		err = viper.SafeWriteConfig()
		if err != nil {
			fmt.Println("Config file already exists, do you want to overwrite it? (yes/no)")
			text, err = reader.ReadString('\n')
			text = strings.TrimSpace(text)
			if err != nil {
				log.Fatal(err)
			}

			if text == "y" || text == "yes" {
				err = viper.WriteConfig()
				if err != nil {
					log.Fatal(err)
				} else {
					log.Println("Config file succesfully written")
				}
			} else {
				log.Println("Write cancelled")
			}
		} else {
			log.Println("Config file succesfully created")
		}
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
