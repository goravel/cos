package facades

import (
	"log"

	"github.com/goravel/framework/contracts/filesystem"

	"github.com/goravel/cos"
)

func Cos(disk string) filesystem.Driver {
	instance, err := cos.App.MakeWith(cos.Binding, map[string]any{"disk": disk})
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*cos.Cos)
}
