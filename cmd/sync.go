/*
Copyright © 2024 Vino <vino0908@outlook.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vinoMamba/notionsync/pkg/notionapi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync data to mongodb",
	Run: func(cmd *cobra.Command, args []string) {
		token := os.Getenv("NOTION_TOKEN")
		client := notionapi.NewClient(notionapi.Token(token))
		databaseId := os.Getenv("NOTION_DATABASE_ID")
		res, err := client.Database.Query(cmd.Context(), notionapi.DatabaseID(databaseId), &notionapi.DatabaseQueryRequest{})
		if err != nil {
			fmt.Println(err)
			return
		}

		uri := os.Getenv("MONGODB_URI")
		c, err := mongo.Connect(cmd.Context(), options.Client().ApplyURI(uri))
		if err != nil {
			fmt.Println(err)
			return
		}

		if _, err := c.Database("blogs").Collection("list").DeleteMany(cmd.Context(), bson.D{}); err != nil {
			fmt.Println(err)
			return
		}
		if _, err := c.Database("blogs").Collection("blocks").DeleteMany(cmd.Context(), bson.D{}); err != nil {
			fmt.Println(err)
			return
		}

		var documents []interface{}
		for _, page := range res.Results {
			documents = append(documents, page)
		}

		if _, err := c.Database("blogs").Collection("list").InsertMany(cmd.Context(), documents); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("sync list success")

		for _, page := range res.Results {
			block, err := client.Block.GetChildren(cmd.Context(), notionapi.BlockID(page.ID), &notionapi.Pagination{PageSize: 100})
			if err != nil {
				fmt.Println(err)
				break
			}
			if _, err := c.Database("blogs").Collection("block").InsertOne(cmd.Context(), block); err != nil {
				fmt.Println(err)
				break
			}
		}
		fmt.Println("sync block success")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
