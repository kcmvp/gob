package common

import "github.com/spf13/cobra"

type ArgFunc func(cmd *cobra.Command) error
