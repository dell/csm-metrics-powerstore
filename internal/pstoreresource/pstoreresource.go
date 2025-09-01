/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
 Copyright (c) 2021-2025 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package pstoreresource

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	csictx "github.com/dell/gocsi/context"

	"github.com/dell/csm-metrics-powerstore/internal/service"
	"github.com/dell/gopowerstore"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

const (
	// EnvThrottlingRateLimit sets a number of concurrent requests to APi
	EnvThrottlingRateLimit = "X_CSI_POWERSTORE_THROTTLING_RATE_LIMIT"
)

// GetPowerStoreArrays parses config.yaml file, initializes gopowerstore Clients and composes map of arrays for ease of access.
// It will return array that can be used as default as a second return parameter.
// If config does not have any array as a default then the first will be returned as a default.
func GetPowerStoreArrays(filePath string, logger *logrus.Logger) (map[string]*service.PowerStoreArray, map[string]string, *service.PowerStoreArray, error) {
	type config struct {
		Arrays []*service.PowerStoreArray `yaml:"arrays"`
	}

	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		logger.WithError(err).Errorf("cannot read file %s", filePath)
		return nil, nil, nil, err
	}

	var cfg config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		logger.WithError(err).Errorf("cannot unmarshal data")
		return nil, nil, nil, err
	}

	arrayMap := make(map[string]*service.PowerStoreArray)
	mapper := make(map[string]string)
	var defaultArray *service.PowerStoreArray
	foundDefault := false

	if len(cfg.Arrays) == 0 {
		return arrayMap, mapper, defaultArray, nil
	}

	// Safeguard if user doesn't set any array as default, we just use first one
	defaultArray = cfg.Arrays[0]

	// Convert to map for convenience and init gopowerstore.Client
	for _, array := range cfg.Arrays {
		array := array
		if array == nil {
			return arrayMap, mapper, defaultArray, nil
		}
		if array.GlobalID == "" {
			return nil, nil, nil, errors.New("no GlobalID field found in config.yaml, update config.yaml according to the documentation")
		}
		clientOptions := gopowerstore.NewClientOptions()
		clientOptions.SetInsecure(array.Insecure)

		if throttlingRateLimit, ok := csictx.LookupEnv(context.Background(), EnvThrottlingRateLimit); ok {
			rateLimit, err := strconv.Atoi(throttlingRateLimit)
			if err != nil {
				logger.Errorf("can't get throttling rate limit, using default")
			} else {
				clientOptions.SetRateLimit(rateLimit) // #nosec G115 -- This is a false positive
			}
		}

		c, err := gopowerstore.NewClientWithArgs(
			array.Endpoint, array.Username, array.Password, clientOptions)
		if err != nil {
			return nil, nil, nil, status.Errorf(codes.FailedPrecondition,
				"unable to create PowerStore client: %s", err.Error())
		}
		array.Client = c
		var ip string
		ips := GetIPListFromString(array.Endpoint)
		if ips == nil {
			logger.Warnf("didn't found an IP from the provided endPoint, it could be a FQDN. Please make sure to enter a valid FQDN in https://abc.com/api/rest format")
			sub := strings.Split(array.Endpoint, "/")
			if len(sub) > 2 {
				ip = sub[2]
				if regexp.MustCompile(`^[0-9.]*$`).MatchString(sub[2]) {
					return nil, nil, nil, fmt.Errorf("can't get ips from endpoint: %s", array.Endpoint)
				}
			} else {
				return nil, nil, nil, fmt.Errorf("can't get ips from endpoint: %s", array.Endpoint)
			}
		} else {
			ip = ips[0]
		}
		array.IP = ip
		logger.Infof("%s,%s,%s,%s,%t,%t,%s", array.Endpoint, array.GlobalID, array.Username, array.NasName, array.Insecure, array.IsDefault, array.BlockProtocol)
		arrayMap[array.GlobalID] = array
		mapper[ip] = array.GlobalID
		if array.IsDefault && !foundDefault {
			defaultArray = array
			foundDefault = true
		}
	}

	return arrayMap, mapper, defaultArray, nil
}

// GetIPListFromString returns list of ips in string form found in input string
// A return value of nil indicates no match
func GetIPListFromString(input string) []string {
	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	return re.FindAllString(input, -1)
}
