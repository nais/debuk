package aivenCmd

import (
	"fmt"
	"github.com/nais/cli/pkg/aiven/secret"
	"github.com/nais/cli/pkg/aiven/services"
	"github.com/urfave/cli/v2"
)

func getCommand() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Aliases:     []string{"g"},
		Description: "Generate preferred config format to '/tmp' folder",
		Before: func(context *cli.Context) error {
			if context.Args().Len() != 3 {
				return fmt.Errorf("missing required arguments: service, secret, namespace")
			}

			_, err := services.ServiceFromString(context.Args().Get(0))
			if err != nil {
				return err
			}
			return nil
		},
		Action: func(context *cli.Context) error {
			service, err := services.ServiceFromString(context.Args().Get(0))
			if err != nil {
				return err
			}

			secretName := context.Args().Get(1)
			namespace := context.Args().Get(2)

			err = secret.ExtractAndGenerateConfig(service, secretName, namespace)
			if err != nil {
				return fmt.Errorf("retrieve secret and generating config: %w", err)
			}

			return nil
		},
	}
}
