package cos

import (
	"context"

	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.cos"

var App foundation.Application

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.BindWith(Binding, func(app foundation.Application, parameter map[string]any) (any, error) {
		return NewCos(context.Background(), app.MakeConfig(), parameter["disk"].(string))
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Publishes("github.com/goravel/cos", map[string]string{
		"config/cos.go": app.ConfigPath(""),
	})
}
