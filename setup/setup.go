package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

var config = `map[string]any{
        "driver": "custom",
        "key":    config.Env("TENCENT_ACCESS_KEY_ID"),
        "secret": config.Env("TENCENT_ACCESS_KEY_SECRET"),
        "url":    config.Env("TENCENT_URL"),
        "via": func() (filesystem.Driver, error) {
            return cosfacades.Cos("cos") // The ` + "`cos`" + ` value is the ` + "`disks`" + ` key
        },
    }`

func main() {
	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&cos.ServiceProvider{}")),
			modify.GoFile(path.Config("filesystems.go")).
				Find(match.Imports()).Modify(modify.AddImport("github.com/goravel/framework/contracts/filesystem"), modify.AddImport("github.com/goravel/cos/facades", "cosfacades")).
				Find(match.Config("filesystems.disks")).Modify(modify.AddConfig("cos", config)).
				Find(match.Config("filesystems")).Modify(modify.AddConfig("default", `"cos"`)),
		).
		Uninstall(
			modify.GoFile(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.Unregister("&cos.ServiceProvider{}")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			modify.GoFile(path.Config("filesystems.go")).
				Find(match.Config("filesystems.disks")).Modify(modify.RemoveConfig("cos")).
				Find(match.Imports()).Modify(modify.RemoveImport("github.com/goravel/framework/contracts/filesystem"), modify.RemoveImport("github.com/goravel/cos/facades", "cosfacades")).
				Find(match.Config("filesystems")).Modify(modify.AddConfig("default", `"local"`)),
		).
		Execute()
}
