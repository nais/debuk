package debug

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

type Debug struct {
	Ctx    context.Context
	Client ctrl.Client
}

func Setup(client ctrl.Client) *Debug {
	return &Debug{
		Ctx:    context.Background(),
		Client: client,
	}
}

func (d *Debug) debug() {
}
