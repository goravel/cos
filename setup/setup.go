package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	config := `map[string]any{
        "driver": "custom",
        "key":    config.Env("TENCENT_ACCESS_KEY_ID"),
        "secret": config.Env("TENCENT_ACCESS_KEY_SECRET"),
        "url":    config.Env("TENCENT_URL"),
        "via": func() (filesystem.Driver, error) {
            return cosfacades.Cos("cos") // The ` + "`cos`" + ` value is the ` + "`disks`" + ` key
        },
    }`

	appConfigPath := path.Config("app.go")
	filesystemsConfigPath := path.Config("filesystems.go")
	moduleImport := setup.Paths().Module().Import()
	cosServiceProvider := "&cos.ServiceProvider{}"
	filesystemContract := "github.com/goravel/framework/contracts/filesystem"
	cosFacades := "github.com/goravel/cos/facades"
	filesystemsDisksConfig := match.Config("filesystems.disks")
	filesystemsConfig := match.Config("filesystems")

	setup.Install(
		// Add cos service provider to app.go if not using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Imports()).Modify(modify.AddImport(moduleImport)).
			Find(match.Providers()).Modify(modify.Register(cosServiceProvider))),

		// Add cos service provider to providers.go if using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.AddProviderApply(moduleImport, cosServiceProvider)),

		// Add cos disk to filesystems.go
		modify.GoFile(filesystemsConfigPath).Find(match.Imports()).Modify(
			modify.AddImport(filesystemContract),
			modify.AddImport(cosFacades, "cosfacades"),
		).
			Find(filesystemsDisksConfig).Modify(modify.AddConfig("cos", config)).
			Find(filesystemsConfig).Modify(modify.AddConfig("default", `"cos"`)),
	).Uninstall(
		// Remove cos disk from filesystems.go
		modify.GoFile(filesystemsConfigPath).
			Find(filesystemsConfig).Modify(modify.AddConfig("default", `"local"`)).
			Find(filesystemsDisksConfig).Modify(modify.RemoveConfig("cos")).
			Find(match.Imports()).Modify(
			modify.RemoveImport(filesystemContract),
			modify.RemoveImport(cosFacades, "cosfacades"),
		),

		// Remove cos service provider from app.go if not using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Providers()).Modify(modify.Unregister(cosServiceProvider)).
			Find(match.Imports()).Modify(modify.RemoveImport(moduleImport))),

		// Remove cos service provider from providers.go if using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.RemoveProviderApply(moduleImport, cosServiceProvider)),
	).Execute()
}
