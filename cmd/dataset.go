package cmd

import (
	"fmt"
	"os"

	"github.com/influxdata/influx-spec/dataset"
	"github.com/influxdata/influx-stress/write"
	"github.com/spf13/cobra"
)

var (
	filterStr string
	host      string
	seed      bool
	teardown  bool
)

func init() {
	datasetCmd := &cobra.Command{
		Use:   "dataset",
		Short: "Run suite of tests to verify that queries return expected results.",
		Run:   runDataset,
	}

	RootCmd.AddCommand(datasetCmd)

	datasetCmd.Flags().StringVarP(&filterStr, "filter", "f", "", "Filter which tests are actually ran.")
	datasetCmd.Flags().StringVarP(&host, "host", "", "http://localhost:8086", "HTTP address for the InfluxDB instance.")
	datasetCmd.Flags().BoolVarP(&seed, "seed", "s", false, "Seed the InfluxDB instance with data.")
	datasetCmd.Flags().BoolVarP(&teardown, "teardown", "t", false, "Drop any databases associated with influx-spec.")
}

func runDataset(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	testDir := args[0]

	datasetDirs := dataset.GetDatasetDirs(testDir)

	cats := dataset.NewSuites(datasetDirs, filterStr)

	for _, cat := range cats {
		cfg := write.ClientConfig{
			BaseURL:   host,
			Database:  fmt.Sprintf("INFLUXDB_SPECIFICATION_TEST_%v", cat.Name()),
			Precision: "s",
		}
		if seed {
			fmt.Printf("Seeding Data for %v\n", cat.Name())
			pointsN, err := cat.Seed(cfg)
			if err != nil {
				fmt.Printf("Encountered Error: %v\nWrote %v points\n", err, pointsN)
				os.Exit(1)
				return
			}
		}

		fmt.Printf("Running Spec for %v\n", cat.Name())
		rs, err := cat.Test(cfg)
		if err != nil {
			fmt.Printf("Encountered Error: %v\n", err)
			os.Exit(1)
			return
		}

		if teardown {
			fmt.Printf("Dropping Database for %v\n", cat.Name())
			err := cat.Teardown(cfg)
			if err != nil {
				fmt.Printf("Encountered Error: %v\n", err)
				os.Exit(1)
				return
			}
		}

		// TODO: add support for other formats
		success := true
		for _, r := range rs {
			if !r.Pass {
				success = false
				fmt.Fprintf(os.Stderr, "Query: %v\nExpected: \n%s\vGot: \n%s\n", r.Description, r.Expected, r.Got)
			}
		}

		if !success {
			os.Exit(1)
		}
	}

}
