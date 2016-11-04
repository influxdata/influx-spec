package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mf metaFlags

func init() {
	metaCmd := &cobra.Command{
		Use:   "meta",
		Short: "Verify that service can perform operations correctly",
		Run:   runMeta,
	}

	RootCmd.AddCommand(metaCmd)

	metaCmd.Flags().BoolVarP(&mf.rp, "rp", "", false, "Run retention policy tests")
	metaCmd.Flags().BoolVarP(&mf.user, "user", "", false, "Run user tests")
}

func runMeta(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		cmd.Usage()
		return
	}

	fmt.Println(mf.rp)
	fmt.Println(mf.user)
}

type metaFlags struct {
	rp   bool
	user bool
}
