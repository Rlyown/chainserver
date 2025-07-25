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
	"path/filepath"
	"encoding/json"
	"fmt"

	"github.com/casibase/chainserver/object"
)


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
	uploadDir := filepath.Join("/home/data/uploads", "strategies", strategyName)
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

	strategyFileItem, err := saveUploadedFile(strategyFileHeader, uploadDir, "strategy")
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
