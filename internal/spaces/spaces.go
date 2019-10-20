// Copyright 2018 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package spaces

import (
	"os"
	"sync"

	"github.com/minio/minio-go"
	log "gopkg.in/clog.v1"

	"github.com/unknwon/gowalker/internal/setting"
)

var client *minio.Client
var clientOnce sync.Once

func Client() *minio.Client {
	clientOnce.Do(func() {
		var err error
		client, err = minio.New(
			setting.DigitalOcean.Spaces.Endpoint,
			setting.DigitalOcean.Spaces.AccessKey,
			setting.DigitalOcean.Spaces.SecretKey,
			true)
		if err != nil {
			log.Fatal(2, "Failed to new minio client: %v", err)
		}
	})
	return client
}

// PutObject uploads an object with given path from local file to the bucket.
func PutObject(localPath, objectName string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	_, err = Client().PutObject(setting.DigitalOcean.Spaces.Bucket, objectName, f, fi.Size(), minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"x-amz-acl": "public-read",
		},
	})
	return err
}

// RemoveObject deletes an object with given path from the bucket.
func RemoveObject(objectName string) error {
	return Client().RemoveObject(setting.DigitalOcean.Spaces.Bucket, objectName)
}
