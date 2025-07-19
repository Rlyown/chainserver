// Copyright 2025 The Casibase Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"fmt"

	"github.com/beego/beego/context"
)

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

func (c *ApiController) ResponseOk(data ...interface{}) {
	resp := Response{Status: "ok"}
	switch len(data) {
	case 2:
		resp.Data2 = data[1]
		fallthrough
	case 1:
		resp.Data = data[0]
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) ResponseError(error string, data ...interface{}) {
	resp := Response{Status: "error", Msg: error}
	switch len(data) {
	case 2:
		resp.Data2 = data[1]
		fallthrough
	case 1:
		resp.Data = data[0]
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) ResponseAudio(audioData []byte, contentType string, filename string) {
	if contentType == "" {
		contentType = "audio/mp3"
	}
	if filename == "" {
		filename = "audio.mp3"
	}

	c.Ctx.Output.Header("Content-Type", contentType)
	c.Ctx.Output.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	err := c.Ctx.Output.Body(audioData)
	if err != nil {
		responseError(c.Ctx, err.Error())
	}
}

func responseError(ctx *context.Context, error string, data ...interface{}) {
	resp := Response{Status: "error", Msg: error}
	switch len(data) {
	case 2:
		resp.Data2 = data[1]
		fallthrough
	case 1:
		resp.Data = data[0]
	}

	err := ctx.Output.JSON(resp, true, false)
	if err != nil {
		panic(err)
	}
}

// Store uploaded file
func saveUploadedFile(fileHeader *multipart.FileHeader, baseDir string, fileType string, uploadType string) (*object.UploadFileItem, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	fileExt := filepath.Ext(fileHeader.Filename)
	newFileName := fileType + fileExt
	savePath := filepath.Join(baseDir, newFileName)

	if err := ioutil.WriteFile(savePath, fileBytes, 0644); err != nil {
		return nil, err
	}

	var url
	if uploadType == "task" {
		url = "/uploads/" + filepath.Join("tasks", filepath.Base(baseDir), newFileName)
	} else {
		url = "/uploads/" + filepath.Join("strategies", filepath.Base(baseDir), newFileName)
	}
	return &object.UploadFileItem{
		Name:        fileHeader.Filename,
		Size:        fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
		URL:         url,
	}, nil
}
