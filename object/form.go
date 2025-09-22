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

package object

type FormItem struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Width string `json:"width"`
}

type Form struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Position    string `xorm:"varchar(100)" json:"position"`

	FormItems []*FormItem `xorm:"varchar(5000)" json:"formItems"`
}

// ==========================
// CT-Sharing related structs
// ==========================
type UploadFileItem struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	URL         string `json:"url"` 
}

type TaskForm struct {
	TaskName	string	`xorm:"varchar(100) notnull pk" json:"taskName" valid:"Required"`
	DataFile    *UploadFileItem `xorm:"json" json:"dataFile"`
	TaskFile    *UploadFileItem	`xorm:"json" json:"taskFile"`
	TaskResult  string           `xorm:"text" json:"taskResult"`
	TaskLog     string           `xorm:"text" json:"taskLog"`
}


// Dataset represents the structure of a dataset object returned by the API
type Dataset struct {
	ID          string `json:"ID"`
	Owner       string `json:"Owner"`
	CreateTime  string `json:"CreateTime"`
	Description string `json:"Description"`
	Digest      string `json:"Digest"`
	ExpireTime  string `json:"ExpireTime"`
	Signature   string `json:"Signature"`
}

// DatasetUsage represents the structure of a dataset usage license returned by the API
type DatasetUsage struct {
	UsageID      string `json:"UsageID"`
	DatasetID    string `json:"DatasetID"`
	User         string `json:"User"`
	ExpireTime   string `json:"ExpireTime"`
	UseCountLeft int    `json:"UseCountLeft"`
}