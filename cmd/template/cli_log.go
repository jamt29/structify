package template

import (
	charmlog "github.com/charmbracelet/log"
	"github.com/jamt29/structify/internal/config"
	"github.com/spf13/cobra"
)

func tmplVerbose(cmd *cobra.Command) bool {
	v, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return false
	}
	return v
}

// tmplStructuredLog returns a logger writing to the command's stderr, for non-TTY stdout.
func tmplStructuredLog(cmd *cobra.Command) *charmlog.Logger {
	l := config.NewLogger(tmplVerbose(cmd))
	l.SetOutput(cmd.ErrOrStderr())
	return l
}
