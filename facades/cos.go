package facades

import (
	"github.com/goravel/framework/contracts/filesystem"

	"github.com/goravel/cos"
)

func Cos(disk string) (filesystem.Driver, error) {
	instance, err := cos.App.MakeWith(cos.Binding, map[string]any{"disk": disk})
	if err != nil {
		return nil, err
	}

	return instance.(*cos.Cos), nil
}
