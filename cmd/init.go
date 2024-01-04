package cmd

import (
	"context"
	"github.com/ArhostCode/devpod-provider-yandexcloud/pkg/yandexcloud"

	"github.com/loft-sh/devpod/pkg/provider"
	"github.com/loft-sh/log"
	"github.com/spf13/cobra"
)

// InitCmd holds the cmd flags
type InitCmd struct{}

// NewInitCmd defines a init
func NewInitCmd() *cobra.Command {
	cmd := &InitCmd{}
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Init account",
		RunE: func(_ *cobra.Command, args []string) error {
			ycProvider, err := yandexcloud.NewProvider(log.Default, true)
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

	return initCmd
}

// Run runs the init logic
func (cmd *InitCmd) Run(
	ctx context.Context,
	ycProvider *yandexcloud.YCProvider,
	machine *provider.Machine,
	logs log.Logger,
) error {
	return yandexcloud.Init(ctx, ycProvider)
}
