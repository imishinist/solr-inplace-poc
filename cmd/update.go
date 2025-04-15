package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/imishinist/solr-inplace-poc/internal/solr"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use: "update",
	RunE: func(cmd *cobra.Command, args []string) error {
		sc := solr.NewClient(solrHost, collection)

		docs := []solr.Document{
			{
				ID: "1",
				Fields: []solr.Field{
					{Key: "x_i", Value: 2},
					{Key: "y_i", Value: 3},
				},
			},
		}
		builder := solr.NewUpdateBatchBuilder(nil, nil)
		builder.Add(docs...)

		var (
			body string
			err  error
		)
		if body, err = builder.Build(); err != nil {
			return err
		}

		resp, err := sc.Update(body)
		if err != nil {
			return err
		}
		defer resp.Close()
		io.Copy(os.Stdout, resp)

		return nil
	},
}

var (
	solrHost   string
	collection string
)

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.PersistentFlags().StringVar(&solrHost, "host", "localhost:8983", "solr host")
	updateCmd.PersistentFlags().StringVar(&collection, "collection", "test", "solr collection")
}
