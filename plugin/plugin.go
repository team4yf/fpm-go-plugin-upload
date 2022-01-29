package plugin

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/team4yf/yf-fpm-server-go/errno"

	"github.com/team4yf/fpm-go-pkg/utils"

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
	Limit        int64
}

type UploadData struct {
	Hash string `json:"hash,omitempty"`
	Path string `json:"path,omitempty"`
}

//UploadRsp the response for upload
type UploadRsp struct {
	URL      string `json:"url"`
	Errno    int    `json:"errno"`
	Uploaded bool   `json:"uploaded"`
	Error    string `json:"error,omitempty"`
	Data     *UploadData
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
				panic(err)
			}
		}

		app.Logger.Debugf("upload Config: %#v", options)

		//Create the upload folder
		_, err := os.Stat(options.Dir)

		if os.IsNotExist(err) {
			errDir := os.MkdirAll(options.Dir, 0755)
			if errDir != nil {
				panic(err)
			}
		}

		app.SetStatic(options.Base, options.Dir)
		app.BindHandler("/download/{filename}", func(c *ctx.Ctx, fpm *fpm.Fpm) {
			filename := c.Param("filename")
			filepath := options.Dir + filename
			file, err := os.Open(filepath)
			if err != nil {
				c.BizError(errno.Wrap(err))
				return
			}
			defer file.Close()
			content, err := ioutil.ReadAll(file)
			if err != nil {
				c.BizError(errno.Wrap(err))
				return
			}
			c.GetResponse().Header().Add("Content-Type", "application/octet-stream")
			c.GetResponse().Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
			c.GetResponse().Write(content)
		}).Methods("GET")
		app.BindHandler(options.UploadRouter, func(c *ctx.Ctx, fpm *fpm.Fpm) {
			r := c.GetRequest()
			r.ParseMultipartForm(32 << 20)
			files := r.MultipartForm.File[options.Field]
			len := len(files)
			fileRspArr := make([]*UploadRsp, len)
			for idx, handler := range files {
				file, err := handler.Open()
				uploadRsp := &UploadRsp{}
				if err != nil {
					uploadRsp.Errno = -1
					uploadRsp.Error = err.Error()
					fileRspArr[idx] = uploadRsp
					continue
				}
				defer file.Close()

				filename := handler.Filename
				ext := path.Ext(filename)
				size := handler.Size
				mime := handler.Header["Content-Type"][0]
				hash := utils.GenShortID()
				dest := options.Dir + hash + ext

				if size > options.Limit<<20 {
					uploadRsp.Errno = -2
					uploadRsp.Error = fmt.Sprintf("upload file should less than %d mb", options.Limit)
					fileRspArr[idx] = uploadRsp
					continue
				}
				accept := false
				for _, m := range options.Accept {
					if m == mime {
						accept = true
					}
				}
				if !accept {
					uploadRsp.Errno = -3
					uploadRsp.Error = "file type not accepted"
					fileRspArr[idx] = uploadRsp
					continue
				}
				f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					uploadRsp.Errno = -4
					uploadRsp.Error = err.Error()
					fileRspArr[idx] = uploadRsp
					continue
				}
				defer f.Close()
				io.Copy(f, file)
				//TODO: calc the files md5

				app.Publish("#file/upload", map[string]string{
					"hash":     hash,
					"filename": filename,
					"ext":      ext,
					"mime":     mime,
					"url":      options.Base + hash + ext,
					"dest":     dest,
					"size":     fmt.Sprintf("%d", size),
				})
				uploadRsp.Errno = 0
				uploadRsp.Uploaded = true
				uploadRsp.URL = options.Base + hash + ext
				uploadRsp.Data = &UploadData{
					Hash: hash,
					Path: options.Base + hash + ext,
				}
				fileRspArr[idx] = uploadRsp
			}
			if len == 1 {
				c.JSON(fileRspArr[0])
				return
			}
			c.JSON(map[string]interface{}{
				"errno":    0,
				"uploaded": true,
				"data":     fileRspArr,
			})

		}).Methods("POST")
	})
}
