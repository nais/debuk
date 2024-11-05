package debugcmd

import (
	ctx "context"
	"flag"
	"fmt"
	"github.com/urfave/cli/v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	debugcmd "k8s.io/kubectl/pkg/cmd/debug"
	"os"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "Debug",
		Usage:           "Launch and attach a debug container",
		ArgsUsage:       "your-pod",
		UsageText:       "nais debug your-pod",
		HideHelpCommand: true,
		Flags:           []cli.Flag{},
		Action: func(context *cli.Context) error {

			var kubeconfig = flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "path to the kubeconfig file")

			// Set up Kubernetes client configuration
			configFlags := genericclioptions.NewConfigFlags(true)
			configFlags.KubeConfig = kubeconfig

			// Set up IO streams (replace os.Stdin, os.Stdout, os.Stderr as needed)
			ioStreams := genericclioptions.IOStreams{
				In:     os.Stdin,
				Out:    os.Stdout,
				ErrOut: os.Stderr,
			}

			// Initialize the debug command with the provided flags and config
			cmd := debugcmd.NewCmdDebug(configFlags, ioStreams)
			cmd.SetArgs([]string{"-it", "amplitrude-proxy-5b6764c656-7nc6v", "--image=\"europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest\"", "--profile=restricted"})

			// Execute the command in context
			if err := cmd.ExecuteContext(ctx.Background()); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to execute kubectl debug: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}
}
