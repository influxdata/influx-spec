package cmd

import (
	"fmt"

	"github.com/influxdata/influx-spec/data"
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
	dataCmd := &cobra.Command{
		Use:   "data",
		Short: "Run suite of tests to verify data ADD BETTER DESCRIPTION.",
		Run:   runData,
	}

	RootCmd.AddCommand(dataCmd)

	dataCmd.Flags().StringVarP(&filterStr, "filter", "f", "", "Run test that match this filter. Better description.")
	dataCmd.Flags().StringVarP(&host, "host", "", "http://localhost:8086", "HTTP address for the InfluxDB instance.")
	dataCmd.Flags().BoolVarP(&seed, "seed", "s", false, "Seed the InfluxDB instance with data.")
	dataCmd.Flags().BoolVarP(&teardown, "teardown", "t", false, "Drop any databases associated with influx-spec.")
}

func runData(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	testDir := args[0]

	dataDirs := data.GetDataDirs(testDir)

	cats := data.NewCategories(dataDirs, filterStr)

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
				return
			}
		}

		fmt.Printf("Running Spec for %v\n", cat.Name())
		err := cat.Test(cfg)
		if err != nil {
			fmt.Printf("Encountered Error: %v\n", err)
			return
		}

		if teardown {
			fmt.Printf("Dropping Database for %v\n", cat.Name())
			err := cat.Teardown(cfg)
			if err != nil {
				fmt.Printf("Encountered Error: %v\n", err)
				return
			}
		}
	}

}
