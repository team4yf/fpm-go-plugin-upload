package main

import (
	"github.com/team4yf/yf-fpm-server-go/fpm"

	_ "github.com/team4yf/fpm-go-plugin-upload/plugin"
)

func main() {

	app := fpm.New()
	app.Init()

	app.Subscribe("#file/upload", func(topic string, data interface{}) {
		app.Logger.Debugf("Receive from upload: %v", data)
	})

	app.Run()

}
