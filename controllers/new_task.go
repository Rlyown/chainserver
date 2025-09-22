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
// @Param datafile1 formData file true "Standard clinical path file"
// @Param datafile2 formData file true "Actual clinical path file"
// @Param taskFile formData file true "Task file"
// @Success 200 {array} object.Form The Response object
// @router /new-task [post]
func (c *ApiController) NewTask() {
	taskName := c.GetString("taskName")

	// Create a directory for the task
	uploadDir := filepath.Join("/home/data/uploads/tasks", taskName)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.ResponseError(fmt.Sprintf("failed to create directory: %s", err.Error()))
		return
	}

	// Handle datafile1 (standard path)
	dataFile1, dataFile1Header, err := c.GetFile("datafile1")
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to get datafile1 (standard path file): %s", err.Error()))
		return
	}
	defer dataFile1.Close()
	datafileExt := filepath.Ext(dataFile1Header.Filename)

	// Handle datafile2 (actual path)
	dataFile2, _, err := c.GetFile("datafile2")
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to get datafile2 (actual path file): %s", err.Error()))
		return
	}
	defer dataFile2.Close()

	dataFile1Bytes, err := io.ReadAll(dataFile1)
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to read datafile1 content: %s", err.Error()))
		return
	}
	dataFile2Bytes, err := io.ReadAll(dataFile2)
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to read datafile2 content: %s", err.Error()))
		return
	}

	// Combine the file contents with a delimiter
	delimiter := []byte("\n==============\n")
	combinedData := bytes.Join([][]byte{dataFile1Bytes, dataFile2Bytes}, delimiter)
	dataFilePath := filepath.Join(uploadDir, "data"+datafileExt)
	err = os.WriteFile(dataFilePath, combinedData, 0644)
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to create combined data file: %s", err.Error()))
		return
	}

	// Write the combined data to a new file
	dataFileItem := &object.UploadFileItem{
		Name:        "data" + datafileExt,
		Size:        int64(len(combinedData)),
		ContentType: dataFile1Header.Header.Get("Content-Type"), // Use the content type of the first file
		URL:         "/files/" + taskName + "/data" + datafileExt,
	}

	// Handle the Wasm task file
	taskFile, taskFileHeader, err := c.GetFile("taskFile")
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
	taskFilePath := filepath.Join(uploadDir, "task"+filepath.Ext(taskFileHeader.Filename))


	// Launch the task
	taskDir := "/home/data/with-chainmaker/CT-Sharing/WASMRuntime_interp/language-bindings/go/samples"
	taskLog, err := launchTask(taskDir, dataFilePath, taskFilePath)
	taskLog = strings.ReplaceAll(taskLog, "ERROR: signal: killed\n", "")
	taskLog = strings.ReplaceAll(taskLog, "ERROR: signal: killed", "")
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to launch task: %s\nTask Output:\n%s", err.Error(), taskLog))
		return
	}
	fmt.Printf(taskLog)

	// Get and parse the task result
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

	// Prepare and send the response
	taskForm := &object.TaskForm{
		TaskName:   taskName,
		DataFile:   dataFileItem, 
		TaskFile:   taskFileItem,
		TaskResult: taskResult,
		TaskLog:    taskLog,
	}
	c.ResponseOk(taskForm)
}
