package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("cos", map[string]any{
		"key":    config.Env("TENCENT_ACCESS_KEY_ID"),
		"secret": config.Env("TENCENT_ACCESS_KEY_SECRET"),
		"bucket": config.Env("TENCENT_BUCKET"),
		"url":    config.Env("TENCENT_URL"),
	})
}
