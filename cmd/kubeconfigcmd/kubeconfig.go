package kubeconfigcmd

import (
	"fmt"
	"github.com/nais/cli/pkg/gcp"
	"github.com/nais/cli/pkg/kubeconfig"
	"github.com/nais/cli/pkg/naisdevice"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "kubeconfig",
		Usage: "Create a kubeconfig file for connecting to available clusters",
		Description: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged
in using:
gcloud auth login --update-adc`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "overwrite",
				Aliases: []string{"o"},
			},
			&cli.BoolFlag{
				Name:    "clear",
				Aliases: []string{"c"},
			},
			&cli.BoolFlag{
				Name:    "include-onprem",
				Aliases: []string{"io"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
		},
		Before: func(context *cli.Context) error {
			err := gcp.ValidateUserLogin(context.Context, false)
			if err != nil {
				return err
			}

			status, err := naisdevice.GetStatus(context.Context)
			if err != nil {
				return err
			}

			if !naisdevice.IsConnected(status) {
				return fmt.Errorf("you need to be connected with naisdevice before using this command")
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			overwrite := context.Bool("overwrite")
			clear := context.Bool("clear")
			includeOnprem := context.Bool("include-onprem")
			verbose := context.Bool("verbose")

			email, err := gcp.GetActiveUserEmail(context.Context)
			if err != nil {
				return err
			}

			return kubeconfig.CreateKubeconfig(context.Context, email, overwrite, clear, includeOnprem, verbose)
		},
	}
}
