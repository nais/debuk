package debugcmd

import (
	"fmt"
	"github.com/nais/cli/pkg/debug"
	"github.com/nais/cli/pkg/k8s"

	"github.com/nais/cli/pkg/metrics"
	"github.com/urfave/cli/v2"
)

// kubectl debug --profile=restricted -it johnnyjob-7hgx4 --image=busybox:1.28

func Command() *cli.Command {
	return &cli.Command{
		Name:      "debug",
		Usage:     "Debug nais application",
		ArgsUsage: "podname",
		Before: func(context *cli.Context) error {
			metrics.AddOne("debug_total")
			if context.Args().Len() < 1 {
				metrics.AddOne("debug_arguments_error_total")
				return fmt.Errorf("missing required arguments: %v", context.Command.ArgsUsage)
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			d := debug.Setup(k8s.SetupControllerRuntimeClient())

			return nil
		},
	}
}
