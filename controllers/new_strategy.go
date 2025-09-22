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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/casibase/chainserver/object"
)

// NewStrategy
// @Title NewStrategy
// @Description Create a new dataset or data usage license
// @Param operationType formData string true "The type of operation: 'createDataset' or 'createDatasetUsage'"
// @Param dataFile formData file false "The data file (only for 'createDataset')"
// @Param datasetId formData string true "The ID of the dataset"
// @Param description formData string false "The description of the dataset"
// @Param owner formData string false "The owner/signature of the dataset"
// @Param expireTime formData string true "The expiration time"
// @Param usageId formData string false "The ID for the usage license (only for 'createDatasetUsage')"
// @Param user formData string false "The user for the usage license (only for 'createDatasetUsage')"
// @Param useCountLeft formData string false "The remaining usage count (only for 'createDatasetUsage')"
// @Success 200 {object} object.Response The Response object
// @router /new-strategy [post]
func (c *ApiController) NewStrategy() {
	operationType := c.GetString("operationType")
	if operationType == "" {
		c.ResponseError("operationType is required")
		return
	}

	// The path to the shell script
	scriptPath := "/home/data/with-chainmaker/CT-Sharing/Strategy/run_strategy.sh" 

	var cmd *exec.Cmd
	var args []string

	switch operationType {
	case "createDataset":
		datasetId := c.GetString("datasetId")
		description := c.GetString("description")
		owner := c.GetString("owner")
		expireTime := c.GetString("expireTime")

		if datasetId == "" || owner == "" || expireTime == "" {
			c.ResponseError("For createDataset, 'datasetId', 'owner', and 'expireTime' are required.")
			return
		}

		dataFile, dataFileHeader, err := c.GetFile("dataFile")
		if err != nil {
			c.ResponseError(fmt.Sprintf("failed to get data file: %s", err.Error()))
			return
		}
		defer dataFile.Close()

		uploadDir := filepath.Join("/home/data/uploads/datasets", datasetId)
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.ResponseError(fmt.Sprintf("failed to create directory: %s", err.Error()))
			return
		}

		_, err = saveUploadedFile(dataFileHeader, uploadDir, "data")
		if err != nil {
			c.ResponseError(fmt.Sprintf("failed to store data file: %s", err.Error()))
			return
		}
		dataFilePath := filepath.Join(uploadDir, "data.json")

		args = []string{scriptPath, "createDataset", datasetId, description, dataFilePath, owner, expireTime}
		cmd = exec.Command(args[0], args[1:]...)

	case "createDatasetUsage":
		usageId := c.GetString("usageId")
		datasetId := c.GetString("datasetId")
		expireTime := c.GetString("expireTime")
		user := c.GetString("user")
		useCountLeft := c.GetString("useCountLeft")

		if usageId == "" || datasetId == "" || expireTime == "" || user == "" || useCountLeft == "" {
			c.ResponseError("For createDatasetUsage, all parameters are required.")
			return
		}
		args = []string{scriptPath, "createDatasetUsage", usageId, datasetId, expireTime, user, useCountLeft}
		cmd = exec.Command(args[0], args[1:]...)

	default:
		c.ResponseError(fmt.Sprintf("Unsupported operationType: %s", operationType))
		return
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		c.ResponseError(fmt.Sprintf("script execution failed: %s\nOutput:\n%s", err.Error(), string(output)))
		return
	}

	fmt.Printf("--- Script Output ---\n%s\n---------------------\n", string(output))

	outputStr := string(output)
	startMarker := "API_RESPONSE_START"
	endMarker := "API_RESPONSE_END"
	startIndex := strings.Index(outputStr, startMarker)
	endIndex := strings.Index(outputStr, endMarker)

	if startIndex == -1 || endIndex == -1 {
		c.ResponseError(fmt.Sprintf("could not find API response markers in script output:\n%s", outputStr))
		return
	}

	jsonStr := outputStr[startIndex+len(startMarker) : endIndex]
	
	switch operationType {
	case "createDataset":
		var datasetResult object.Dataset
		if err := json.Unmarshal([]byte(jsonStr), &datasetResult); err != nil {
			c.ResponseError(fmt.Sprintf("failed to parse Dataset result from script: %s\nJSON String:\n%s", err.Error(), jsonStr))
			return
		}
		c.ResponseOk(datasetResult)
	case "createDatasetUsage":
		var usageResult object.DatasetUsage
		if err := json.Unmarshal([]byte(jsonStr), &usageResult); err != nil {
			c.ResponseError(fmt.Sprintf("failed to parse DatasetUsage result from script: %s\nJSON String:\n%s", err.Error(), jsonStr))
			return
		}
		c.ResponseOk(usageResult)
	}
}