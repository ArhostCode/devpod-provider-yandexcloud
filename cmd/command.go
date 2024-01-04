package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ArhostCode/devpod-provider-yandexcloud/pkg/yandexcloud"
	"github.com/loft-sh/devpod/pkg/provider"
	"github.com/loft-sh/devpod/pkg/ssh"
	"github.com/loft-sh/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CommandCmd holds the cmd flags
type CommandCmd struct{}

// NewCommandCmd defines a command
func NewCommandCmd() *cobra.Command {
	cmd := &CommandCmd{}
	commandCmd := &cobra.Command{
		Use:   "command",
		Short: "Command an instance",
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

	return commandCmd
}

// Run runs the command logic
func (cmd *CommandCmd) Run(
	ctx context.Context,
	ycProvider *yandexcloud.YCProvider,
	machine *provider.Machine,
	logs log.Logger,
) error {
	command := os.Getenv("COMMAND")

	if command == "" {
		return fmt.Errorf("command environment variable is missing")
	}

	//log.GetInstance().Info("Loading key")

	privateKey, err := ssh.GetPrivateKeyRawBase(ycProvider.Config.MachineFolder)

	if err != nil {
		return fmt.Errorf("load private key: %w", err)
	}

	//log.GetInstance().Info("Private key loaded")

	// get instance

	//log.GetInstance().Info("Getting Instance")

	instance, err := yandexcloud.GetDevpodInstance(ctx, ycProvider)
	if err != nil {
		return err
	}

	//log.GetInstance().Info("Getting Interfaces")
	ip := instance.NetworkInterfaces[0].PrimaryV4Address.OneToOneNat.Address
	//log.GetInstance().Infof("IP %v\n", ip)
	sshClient, err := ssh.NewSSHClient("devpod", ip+":22", privateKey)

	if err != nil {
		return errors.Wrap(err, "create ssh client")
	}

	defer sshClient.Close()

	// run command
	return ssh.Run(ctx, sshClient, command, os.Stdin, os.Stdout, os.Stderr)
}
