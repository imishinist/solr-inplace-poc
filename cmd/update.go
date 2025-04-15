package cmd

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/imishinist/solr-inplace-poc/internal/solr"
)

func parseCSV(in io.Reader) ([]solr.Document, error) {
	reader := csv.NewReader(in)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// extract "ID"
	idIndex := -1
	for i, field := range header {
		if strings.ToUpper(field) == "ID" {
			idIndex = i
		}
	}
	if idIndex == -1 {
		return nil, errors.New("csv should contains id field")
	}

	docs := make([]solr.Document, 0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		doc := solr.Document{
			Fields: make([]solr.Field, 0, len(record)),
		}
		for i, field := range record {
			if i == idIndex {
				doc.ID = field
				continue
			}
			doc.Fields = append(doc.Fields, solr.Field{
				Key:   header[i],
				Value: field,
			})
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func openFile(fileName string) (io.Reader, error) {
	if fileName == "-" {
		return os.Stdin, nil
	}
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, f)
	return buf, nil
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use: "update",
	RunE: func(cmd *cobra.Command, args []string) error {
		sc := solr.NewClient(solrHost, collection)

		if csvFile == "" {
			return errors.New("csv file is empty")
		}

		in, err := openFile(csvFile)
		if err != nil {
			return err
		}

		docs, err := parseCSV(in)
		if err != nil {
			return err
		}

		builder := solr.NewUpdateBatchBuilder(nil, nil)
		builder.Add(docs...)

		var (
			body string
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

	csvFile string
)

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.PersistentFlags().StringVar(&solrHost, "host", "localhost:8983", "solr host")
	updateCmd.PersistentFlags().StringVar(&collection, "collection", "test", "solr collection")

	updateCmd.PersistentFlags().StringVar(&csvFile, "csv", "-", "csv file")
}
