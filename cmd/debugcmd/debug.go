package debugcmd

import (
	"fmt"
	"github.com/nais/cli/pkg/debug"
	"github.com/nais/cli/pkg/k8s"
	"github.com/nais/cli/pkg/metrics"
	"github.com/urfave/cli/v2"
)

// kubectl debug --profile=restricted -it johnnyjob-7hgx4 --image=busybox:1.28
const (
	namespaceFlagName = "namespace"
	contextFlagName   = "context"
	debugImageDefault = "europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "debug",
		Usage:     "Debug nais application",
		ArgsUsage: "podname",
		Flags: []cli.Flag{
			namespaceFlag(),
			kubeConfigFlag(),
		},
		Before: func(context *cli.Context) error {
			metrics.AddOne("debug_total")
			if context.Args().Len() < 1 {
				metrics.AddOne("debug_arguments_error_total")
				return fmt.Errorf("missing required arguments: %v", context.Command.ArgsUsage)
			}

			return nil
		},
		Action: func(cCtx *cli.Context) error {
			cfg := makeConfig(cCtx)
			cluster := cCtx.String(contextFlagName)
			namespace := cCtx.String(namespaceFlagName)
			client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))
			cfg.Namespace = client.CurrentNamespace
			if namespace != "" {
				cfg.Namespace = namespace
			}

			clientset, err := k8s.SetupClientGo(cluster)
			if err != nil {
				return err
			}
			dg := debug.Setup(clientset, cfg)
			p, err := dg.Debug()
			if err != nil {
				return fmt.Errorf("error debugging instance: %w", err)
			}
			err = dg.AttachToEphemeralContainer(p)
			if err != nil {
				return fmt.Errorf("error attaching to ephemeral container: %w", err)
			}
			return nil
		},
	}
}

func namespaceFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        namespaceFlagName,
		DefaultText: "The namespace from your current kubeconfig context",
		Usage:       "The kubernetes `NAMESPACE` to use",
		Aliases:     []string{"n"},
	}
}

func kubeConfigFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        contextFlagName,
		Aliases:     []string{"c"},
		Usage:       "The kubeconfig `CONTEXT` to use",
		DefaultText: "The current context in your kubeconfig",
	}
}

func makeConfig(cCtx *cli.Context) debug.Config {
	podName := cCtx.Args().Get(0)
	namespace := cCtx.Args().Get(1)

	return debug.Config{
		PodName:    podName,
		Namespace:  namespace,
		DebugImage: debugImageDefault,
	}
}
