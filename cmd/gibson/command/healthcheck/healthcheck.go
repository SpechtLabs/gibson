// Package healthcheck implements gibson healthcheck, which checks that a
// gibson serve answers.
package healthcheck

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/spechtlabs/gibson/internal/config"
	"github.com/spechtlabs/gibson/internal/server"
)

// NewCommand returns the healthcheck command, which reads the address from
// v's ssh.listen, the key serve listens on.
func NewCommand(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Exit 0 if the SSH server answers",
		Long: `Exits 0 if something on --listen answers with an SSH banner, and 1 otherwise.
The container image has no shell, so its HEALTHCHECK runs the binary itself.
An address without a host is checked on 127.0.0.1.`,
		Example: `gibson healthcheck --listen :2222`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := config.Bind(v, cmd.Flags(), map[string]string{"listen": config.Listen}); err != nil {
				return err
			}
			if err := server.Healthcheck(v.GetString(config.Listen)); err != nil {
				return fmt.Errorf("unhealthy: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().String("listen", ":2222", "address the SSH server listens on")
	return cmd
}
