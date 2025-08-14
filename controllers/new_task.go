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
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"fmt"
	"io"
	"strings"
	"github.com/casibase/chainserver/object"
)


// Launch CT-Sharing task
func launchTask(taskDir string, dataFilePath string, taskFilePath string) (string, error) {
	scriptPath := "/home/data/with-chainmaker/CT-Sharing/WASMRuntime_interp/language-bindings/go/samples/start.sh"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", err
	}
	
	fmt.Printf("scriptPath: %s\n", scriptPath)
    	fmt.Printf("dataFilePath: %s\n", dataFilePath)
    	fmt.Printf("taskFilePath: %s\n", taskFilePath)
	
	/*
	dataFilePath := filepath.Join(taskDir, "data"+datafileExt)
	taskFilePath := filepath.Join(taskDir, "task"+taskfileExt)
	*/

	cmd := exec.Command(scriptPath, dataFilePath, taskFilePath)
	cmd.Dir = taskDir
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr

	var outputBuf bytes.Buffer
	//cmd.Stdout = &outputBuf
    	//cmd.Stderr = &outputBuf
	
	cmd.Stdout = io.MultiWriter(os.Stdout, &outputBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &outputBuf)

	fmt.Printf("Executing command in dir %s: %s\n", cmd.Dir, cmd.String())
	err := cmd.Run()
	/*
	if err != nil {
		return err
	}

	return nil
	*/
	
	fmt.Println(outputBuf.String())
	fmt.Println("error:", err)

	return outputBuf.String(), err
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
	taskName := c.GetString("taskName")
	//cryptoPath := c.GetString("cryptoPath")

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
	if err != nil {
                c.ResponseError(fmt.Sprintf("failed to get task file: %s", err.Error()))
                return
        }
	taskfileExt := filepath.Ext(taskFileHeader.Filename)
	
	defer taskFile.Close()

	taskFileItem, err := saveUploadedFile(taskFileHeader, uploadDir, "task")
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to store task file: %s", err.Error()))
		return
	}

	// launch CT-Sharing task
	taskDir := "/home/data/with-chainmaker/CT-Sharing/WASMRuntime_interp/language-bindings/go/samples"
	
	dataFilePath := filepath.Join(uploadDir, "data"+datafileExt)
	taskFilePath := filepath.Join(uploadDir, "task"+taskfileExt)
	
	// get task log
    taskLog, err := launchTask(taskDir, dataFilePath, taskFilePath)
	taskLog = strings.ReplaceAll(taskLog, "ERROR: signal: killed\n", "")
	taskLog = strings.ReplaceAll(taskLog, "ERROR: signal: killed", "")
	if err != nil {
	    	c.ResponseError(fmt.Sprintf("failed to launch task: %s\nTask Output:\n%s", err.Error(), taskLog))
	    	return
	}
	
	fmt.Printf(taskLog)

    // get task result
    resultDir := "/home/data/with-chainmaker/CT-Sharing/DataUser/"
	resultFile := filepath.Join(resultDir, "output_datauser.log")
	
	resultContent, err := os.ReadFile(resultFile)
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to read result file: %s", err.Error()))
		return
	}
	
	resultLog := string(resultContent)
    startMarker := "result:"
    endMarker := "time3 starts with get the du_ResultPackage and ends with get the result"
    var taskResult string
    if s := strings.Index(resultLog, startMarker); s >= 0 {
        s += len(startMarker)
        if e := strings.Index(resultLog[s:], endMarker); e >= 0 {
            taskResult = strings.TrimSpace(resultLog[s : s+e])
        } else {
            taskResult = strings.TrimSpace(resultLog[s:])
        }
    }

	taskForm := &object.TaskForm{
		TaskName:   taskName,
		DataFile:   dataFileItem,
		TaskFile:   taskFileItem,
        TaskResult: taskResult,
		TaskLog:    taskLog,
	}
	c.ResponseOk(taskForm)
}
