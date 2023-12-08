package cmd

import (
	"github.com/kcmvp/gob/cmd/common"
	"github.com/spf13/cobra"
)

var onionFunc common.ArgFunc = func(cmd *cobra.Command) error {
	return nil
}
