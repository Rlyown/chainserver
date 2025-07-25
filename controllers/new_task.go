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
	"os"
	// "os/exec"
	"path/filepath"
	// "encoding/json"
	"fmt"
	// "strings"

	"github.com/casibase/chainserver/object"
)


// Launch CT-Sharing task
func launchTask(taskDir string, datafileExt string, taskfileExt string) error {
	// TODO(shejiarui): hard code here, modify it in test environment
	scriptPath := "/home/data/with-chainmaker/CT-Sharing/WASMRuntime_interp/language-bindings/go/samples/start.sh"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return err
	}

	dataFilePath := filepath.Join(taskDir, "data"+datafileExt)
	taskFilePath := filepath.Join(taskDir, "task"+taskfileExt)

	fmt.Printf("scriptPath: %s, dataFilePath: %s, taskFilePath: %s\n", 
				scriptPath, dataFilePath, taskFilePath)

	// cmd := exec.Command(scriptPath, dataFilePath, taskFilePath)
	// cmd.Dir = taskDir

	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	return err
	// }

	// outputStr := strings.TrimSpace(string(output))
	// if outputStr != "" {
	// 	fmt.Printf("Task info: %s\n", outputStr)
	// }

	return nil
}


// NewTask
// @Title NewTask
// @Description invoke CT-Sharing to run a new task
// @Param taskName formData string true "The name of task"
// @Param secretKey formData string true "The secret key of data"
// @Param dataFile formData file true "Data file"
// @Param taskFile formData file true "Task file"
// @Success 200 {array} object.Form The Response object
// @router /new-task [post]
func (c *ApiController) NewTask() {
	// var taskFormObj object.TaskForm
	// err := json.Unmarshal(c.Ctx.Input.RequestBody, &taskFormObj)
	// if err != nil {
	// 	c.ResponseError(err.Error())
	// 	return
	// }

	taskName := c.GetString("taskName")
	secretKey := c.GetString("secretKey")

	// create directory to store uploaded file
	uploadDir := filepath.Join("/home/data/uploads/tasks", taskName)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.ResponseError(fmt.Sprintf("failed to create directory: %s", err.Error()))
		return
	}

	// store uploaded data file
	dataFile, dataFileHeader, err := c.GetFile("dataFile")
	datafileExt := filepath.Ext(dataFileHeader.Filename)
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to get data file: %s", err.Error()))
		return
	}
	defer dataFile.Close()

	dataFileItem, err := saveUploadedFile(dataFileHeader, uploadDir, "data")
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to store data file: %s", err.Error()))
		return
	}

	// store uploaded task file
	taskFile, taskFileHeader, err := c.GetFile("taskFile")
	taskfileExt := filepath.Ext(taskFileHeader.Filename)
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to get task file: %s", err.Error()))
		return
	}
	defer taskFile.Close()

	taskFileItem, err := saveUploadedFile(taskFileHeader, uploadDir, "task")
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to store task file: %s", err.Error()))
		return
	}

	// launch CT-Sharing task
	if err := launchTask(uploadDir, datafileExt, taskfileExt); err != nil {
		c.ResponseError(fmt.Sprintf("failed to launch task: %s", err.Error()))
		return
	}

	taskForm := &object.TaskForm{
		TaskName:  taskName,
		SecretKey: secretKey,
		DataFile:  dataFileItem,
		TaskFile:  taskFileItem,
	}
	c.ResponseOk(taskForm)
}
