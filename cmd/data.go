package cmd

import (
	"fmt"

	"github.com/influxdata/influx-spec/data"
	"github.com/influxdata/influx-stress/write"
	"github.com/spf13/cobra"
)

var (
	filterStr string
)

func init() {
	dataCmd := &cobra.Command{
		Use:   "data",
		Short: "Run suite of tests to verify data ADD BETTER DESCRIPTION.",
		Run:   runData,
	}

	RootCmd.AddCommand(dataCmd)

	dataCmd.Flags().StringVarP(&filterStr, "filter", "f", "", "Run test that match this filter. Better description")
}

func runData(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	testDir := args[0]

	dataDirs := data.GetDataDirs(testDir, filterStr)

	cats := data.NewCategories(dataDirs)

	for _, cat := range cats {
		fmt.Println(cat.Name())
		cfg := write.ClientConfig{
			BaseURL:   "http://localhost:8086",
			Database:  cat.Name(),
			Precision: "s",
		}
		pointsN, err := cat.Seed(cfg)
		if err != nil {
			fmt.Println(err)
			fmt.Println(pointsN)
			return
		}
		err = cat.Test(cfg)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

}
