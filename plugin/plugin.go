package plugin

import (
	"fmt"
	"io"
	"os"

	"github.com/team4yf/yf-fpm-server-go/ctx"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

//Options options of upload
type Options struct {
	Dir          string
	Field        string
	UploadRouter string
	Base         string
	Accept       []string
	Limit        int
}

func init() {
	fpm.Register(func(app *fpm.Fpm) {
		// 配置 上传
		options := Options{
			Dir:          "public/uploads/",
			Field:        "upload",
			UploadRouter: "/upload",
			Base:         "/uploads/",
			Accept: []string{
				"application/octet-stream",
				"application/json",
				"application/zip",
				"application/x-zip-compressed",
				"image/png",
				"image/jpeg",
			},
			Limit: 5,
		}
		if app.HasConfig("upload") {
			if err := app.FetchConfig("upload", &options); err != nil {
				panic(fmt.Errorf("Load Upload Config Error: %v", err))
			}
		}

		app.Logger.Debugf("upload Config: %#v", options)

		_, err := os.Stat(options.Dir)

		if os.IsNotExist(err) {
			errDir := os.MkdirAll(options.Dir, 0755)
			if errDir != nil {
				panic(err)
			}

		}

		app.BindHandler(options.UploadRouter, func(c *ctx.Ctx, fpm *fpm.Fpm) {
			r := c.GetRequest()
			r.ParseMultipartForm(32 << 20)
			file, handler, err := r.FormFile(options.Field)
			if err != nil {
				c.JSON(map[string]interface{}{
					"errno":    -1,
					"uploaded": false,
					"error":    err,
				})
				return
			}
			defer file.Close()

			f, err := os.OpenFile(options.Dir+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666) // 此处假设当前目录下已存在test目录
			if err != nil {
				c.JSON(map[string]interface{}{
					"errno":    -2,
					"uploaded": false,
					"error":    err,
				})
				return
			}
			defer f.Close()
			io.Copy(f, file)
			c.JSON(map[string]interface{}{
				"errno":    0,
				"uploaded": true,
				"url":      "http://cdn.yunplus.io/408e19877d7b2c73_test.json",
				"data": map[string]interface{}{
					"hash": "408e19877d7b2c73_test",
					"path": "http://cdn.yunplus.io/408e19877d7b2c73_test.json",
				},
			})
		}).Methods("POST")

		bizModule := make(fpm.BizModule, 0)

		bizModule["send"] = func(param *fpm.BizParam) (data interface{}, err error) {
			data = 1
			return
		}

		app.AddBizModule("upload", &bizModule)

	})
}
