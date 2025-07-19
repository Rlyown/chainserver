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
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"encoding/json"

	"github.com/casibase/chainserver/object"
)


// Store uploaded file at '/uploads/tasks/${taskName}'
func saveUploadedFile(fileHeader *multipart.FileHeader, baseDir string, fileType string) (*object.UploadFileItem, error) {
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

	return &object.UploadFileItem{
		Name:        fileHeader.Filename,
		Size:        fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
		URL:         "/uploads/" + filepath.Join("tasks", filepath.Base(baseDir), newFileName),
	}, nil
}


// Launch CT-Sharing task
func (c *FileUploadController) launchTask(taskDir string, taskForm *object.TaskForm) error {
	// TODO(shejiarui): hard code here, modify it in test environment
	scriptPath := "/home/daqi/with-log/CT-Sharing/WASMRuntime_interp/language-bindings/go/samples/start.sh"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return err
	}

	dataFilePath := filepath.Join(taskDir, "data"+filepath.Ext(taskForm.DataFile.Name))
	taskFilePath := filepath.Join(taskDir, "task"+filepath.Ext(taskForm.TaskFile.Name))

	cmd := exec.Command(scriptPath, dataFilePath, taskFilePath)
	cmd.Dir = taskDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != "" {
		web.BeeLogger.Info("Task info: %s", outputStr)
	}

	return nil
}


// NewStrategy
// @Title NewStrategy
// @Description upload a new data usage strategy
// @Param strategyName formData string true "The name of strategy"
// @Param strategyFile formData file true "Strategy file"
// @Success 200 {array} object.Form The Response object
// @router /new-strategy [post]
func (c *ApiController) NewStrategy() {
	var strategyFormObj object.StrategyForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &strategyFormObj)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	strategyName := c.GetString("strategyName")

	// create directory to store uploaded file
	uploadDir := filepath.Join("uploads", "strategies", strategyName)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.ResponseError(fmt.Sprintf("failed to create directory: %s", err.Error()))
		return
	}

	// store uploaded strategy file
	strategyFile, strategyFileHeader, err := c.GetFile("strategyFile")
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to get strategy file: %s", err.Error()))
		return
	}
	defer strategyFile.Close()

	strategyFileItem, err := saveUploadedFile(strategyFileHeader, uploadDir, "strategy", "strategy")
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to store strategy file: %s", err.Error()))
		return
	}

	strategyForm := &object.StrategyForm{
		StrategyName:  strategyName,
		StrategyFile:  strategyFileItem,
	}
	c.ResponseOk(strategyForm)
}
