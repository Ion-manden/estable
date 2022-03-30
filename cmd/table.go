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
	"io"
	"log"
	"os"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fields []string

// tableCmd represents the table command
var tableCmd = &cobra.Command{
	Use:   "table",
	Short: "Create a table of es data from doc ids from stdin",
	Long: `Creates a table view of data from elasticsearch.

  Doc ids are provided by stdin
  "cat es_doc_ids.csv | estable table"

  Fields are specified with the --field flag, multiple fields can be used by setting --field multiple times
  "--field name --field age"`,
	Run: func(cmd *cobra.Command, args []string) {
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

		app := tview.NewApplication()
		table := tview.NewTable().SetBorders(true)

		for ci, col := range fields {
			table.SetCell(
				0,
				ci,
				tview.NewTableCell(strings.Replace(col, "_source.", "", 1)).
					SetTextColor(tcell.ColorYellow).
					SetAlign(tview.AlignCenter),
			)
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

			for ci, field := range fields {
				cellText, ok := doc.Path(field).Data().(string)
				if !ok {
					cellText = ""
				}

				if len(cellText) > 50 {
					cellText = fmt.Sprint(cellText[:50], "...")
				}

				color := tcell.ColorWhite
				if ci == 0 {
					if doc.Path("found").Data().(bool) {
						color = tcell.ColorGreen
					} else {
						color = tcell.ColorRed
					}
				}

				table.SetCell(
					ri,
					ci,
					tview.NewTableCell(cellText).
						SetTextColor(color).
						SetAlign(tview.AlignCenter),
				)
			}
		}

		table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				app.Stop()
			}
			if key == tcell.KeyEnter {
				table.SetSelectable(true, true)
			}
		}).SetSelectedFunc(func(row int, column int) {
			table.GetCell(row, column).SetTextColor(tcell.ColorRed)
			table.SetSelectable(false, false)
		})
		if err := app.SetRoot(table, true).EnableMouse(true).Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tableCmd)

	// Here you will define your flags and configuration settings.
	tableCmd.Flags().StringArrayVarP(&fields, "field", "f", []string{}, "Field to show in table")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tableCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tableCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
