package cmd

import (
	"context"
	"github.com/ArhostCode/devpod-provider-yandexcloud/pkg/yandexcloud"

	"github.com/loft-sh/devpod/pkg/provider"
	"github.com/loft-sh/log"
	"github.com/spf13/cobra"
)

// StartCmd holds the cmd flags
type StartCmd struct{}

// NewStartCmd defines a command
func NewStartCmd() *cobra.Command {
	cmd := &StartCmd{}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			ycProvider, err := yandexcloud.NewProvider(log.Default, false)
			if err != nil {
				return err
			}

			return cmd.Run(
				context.Background(),
				ycProvider,
				provider.FromEnvironment(),
				log.Default,
			)
		},
	}

	return startCmd
}

// Run runs the command logic
func (cmd *StartCmd) Run(
	ctx context.Context,
	ycProvider *yandexcloud.YCProvider,
	machine *provider.Machine,
	logs log.Logger,
) error {
	return yandexcloud.Start(ctx, ycProvider)
}
