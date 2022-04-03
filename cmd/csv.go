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
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var outFile string

// csvCmd represents the csv command
var csvCmd = &cobra.Command{
	Use:   "csv",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("csv called")

		address := viper.GetString("es_address")
		user := viper.GetString("es_user")
		password := viper.GetString("es_password")
		index := viper.GetString("es_index")

		es, err := elasticsearch.NewClient(elasticsearch.Config{
			Username: user,
			Password: password,
			Addresses: []string{
				address,
			},
		})
		if err != nil {
			log.Fatalf("Error creating the client: %s", err)
		}

		reader := bufio.NewReader(os.Stdin)

		if inputFile != "" {
			file, err := os.Open(inputFile)
			if err != nil {
				log.Fatalf("error opening file: %s", err)
			}
			defer file.Close()

			reader = bufio.NewReader(file)
		}

		rawFields := fields
		fields = []string{"_id"}
		for _, rf := range rawFields {
			fields = append(fields, fmt.Sprint("_source.", rf))
		}

		body := gabs.New()

		for {
			text, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			text = strings.Trim(text, "\r\n")

			obj := gabs.New()
			obj.SetP(index, "_index")
			obj.SetP(text, "_id")

			err = body.ArrayAppend(obj, "docs")
			if err != nil {
				log.Fatalf("error appending to array: %s", err)
			}
		}

		res, err := es.Mget(strings.NewReader(body.String()), es.Mget.WithSourceIncludes(rawFields...), es.Mget.WithIndex(index))
		if err != nil {
			log.Fatalf("error calling mget: %s", err)
		}
		defer res.Body.Close()
		if res.IsError() {
			log.Fatalf("Error: %s", res.String())
		}

		w := csv.NewWriter(bufio.NewWriter(os.Stdout))

		if outFile != "" {
			file, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				log.Fatalf("error opening file: %s", err)
			}
			defer file.Close()

			w = csv.NewWriter(bufio.NewWriter(file))
		}

		err = w.Write(fields)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("error reading body: %s", err)
		}
		parsedJson, err := gabs.ParseJSON(b)
		if err != nil {
			log.Fatalf("error parsing json: %s", err)
		}

		for ri, doc := range parsedJson.Search("docs").Children() {
			ri++

			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			cells := []string{}
			for _, field := range fields {
				cellText, ok := doc.Path(field).Data().(string)
				if !ok {
					cellText = ""
				}

				cells = append(cells, cellText)
			}

			err = w.Write(cells)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(csvCmd)

	// Here you will define your flags and configuration settings.
	csvCmd.Flags().StringVar(&outFile, "out", "", "File to save output instead of stdin")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// csvCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// csvCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
