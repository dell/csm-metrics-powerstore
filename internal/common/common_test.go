/*
 Copyright (c) 2021-2022 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package common_test

import (
	"context"
	"testing"

	"github.com/dell/csm-metrics-powerstore/internal/common"
	csictx "github.com/dell/gocsi/context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_Run(t *testing.T) {
	tests := map[string]func(t *testing.T) (filePath string, env map[string]string, expectError bool){
		"success": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/sample-config.yaml", map[string]string{common.EnvThrottlingRateLimit: "123"}, false
		},
		"invalid throttling value": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/sample-config.yaml", map[string]string{common.EnvThrottlingRateLimit: "abc"}, false
		},
		"file doesn't exist": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/no-file.yaml", nil, true
		},
		"file format": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/invalid-format.yaml", nil, true
		},
		"no global id": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/no-global-id.yaml", nil, true
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			logger := logrus.New()
			filePath, envs, expectError := test(t)

			for k, v := range envs {
				csictx.Setenv(context.Background(), k, v)
			}

			arrays, mapper, defaultArray, err := common.GetPowerStoreArrays(filePath, logger)

			if expectError {
				assert.Nil(t, arrays)
				assert.Nil(t, mapper)
				assert.Nil(t, defaultArray)
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, arrays)
				assert.NotNil(t, mapper)
				assert.NotNil(t, defaultArray)
				assert.Nil(t, err)
			}
		})
	}
}
