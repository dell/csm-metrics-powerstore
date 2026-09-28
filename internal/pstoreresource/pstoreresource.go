/*
 Copyright (c) 2021-2026 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	"net"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/dell/csm-metrics-powerstore/internal/service"
	"github.com/dell/csmlog"
	csictx "github.com/dell/gocsi/context"
	"github.com/dell/gopowerstore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

var lookupIPFunc = func(ctx context.Context, host string) ([]netip.Addr, error) {
	return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
}

const (
	// Default timeout for powerstore API call
	defaultAPITimeout = 120 * time.Second

	// EnvThrottlingRateLimit sets a number of concurrent requests to APi
	EnvThrottlingRateLimit = "X_CSI_POWERSTORE_THROTTLING_RATE_LIMIT"

	// EnvPowerstoreAPITimeout specifies the timeout for Powerstore REST API calls
	EnvPowerstoreAPITimeout = "X_CSI_POWERSTORE_API_TIMEOUT"
)

// GetPowerStoreArrays parses config.yaml file, initializes gopowerstore Clients and composes map of arrays for ease of access.
// It will return array that can be used as default as a second return parameter.
// If config does not have any array as a default then the first will be returned as a default.
func GetPowerStoreArrays(filePath string) (map[string]*service.PowerStoreArray, map[string]string, *service.PowerStoreArray, error) {
	type config struct {
		Arrays []*service.PowerStoreArray `yaml:"arrays"`
	}

	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		csmlog.WithFields(csmlog.Fields{
			"error": err,
			"file":  filePath,
		}).Error("cannot read file")
		return nil, nil, nil, err
	}

	var cfg config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		csmlog.WithFields(csmlog.Fields{
			"error": err,
		}).Error("cannot unmarshal data")
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

		// Set timeout for powerstore API call
		timeout := defaultAPITimeout
		if powerStoreAPITimeout, ok := csictx.LookupEnv(context.Background(), EnvPowerstoreAPITimeout); ok {
			fetchedTimeout, err := time.ParseDuration(powerStoreAPITimeout)
			if err != nil {
				csmlog.Errorf("can't get api timeout, using default. error : %s", err)
			} else {
				timeout = fetchedTimeout
				csmlog.Infof("%s set to: %v", EnvPowerstoreAPITimeout, timeout)
			}
		}
		clientOptions.SetDefaultTimeout(timeout)

		if throttlingRateLimit, ok := csictx.LookupEnv(context.Background(), EnvThrottlingRateLimit); ok {
			rateLimit, err := strconv.Atoi(throttlingRateLimit)
			if err != nil {
				csmlog.Errorf("can't get throttling rate limit, using default")
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
		var networkProtocol string
		ips := GetIPListFromString(array.Endpoint)
		if ips == nil {
			csmlog.Warnf("didn't found an IP from the provided endPoint, it could be a FQDN. Please make sure to enter a valid FQDN in https://abc.com/api/rest format")
			// FR-11.2: Use url.Parse().Hostname() to correctly strip brackets from IPv6 addresses
			u, err := url.Parse(array.Endpoint)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("can't parse endpoint: %s", array.Endpoint)
			}
			ip = u.Hostname()
			// Validate that hostname is not just an IP address (which should have been caught above)
			if regexp.MustCompile(`^[0-9.]*$`).MatchString(ip) {
				return nil, nil, nil, fmt.Errorf("can't get ips from endpoint: %s", array.Endpoint)
			}

			resolveCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			resolved, resolveErr := lookupIPFunc(resolveCtx, ip)
			cancel()
			if resolveErr == nil {
				for _, addr := range resolved {
					family := InferProtocol(addr.String())
					if family == "unknown" {
						continue
					}
					if networkProtocol == "" {
						networkProtocol = family
					} else if networkProtocol != family {
						networkProtocol = "unknown"
						break
					}
				}
			}
		} else {
			ip = ips[0]
			networkProtocol = InferProtocol(ip)
		}
		array.IP = ip
		array.NetworkProtocol = networkProtocol
		csmlog.Infof("%s,%s,%s,%s,%t,%t,%s", array.Endpoint, array.GlobalID, array.Username, array.NasName, array.Insecure, array.IsDefault, array.BlockProtocol)
		arrayMap[array.GlobalID] = array
		mapper[ip] = array.GlobalID
		if array.IsDefault && !foundDefault {
			defaultArray = array
			foundDefault = true
		}
	}

	return arrayMap, mapper, defaultArray, nil
}

// InferProtocol determines the IP protocol (ipv4 or ipv6) from an address string
// FR-10.1: Helper function to classify IP addresses for metrics labeling
func InferProtocol(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return "unknown"
	}

	if addr.Is4() {
		return "ipv4"
	}
	if addr.Is6() && !addr.Is4In6() {
		return "ipv6"
	}
	if addr.Is4In6() {
		// IPv4-mapped IPv6 addresses should be classified as ipv6
		return "ipv6"
	}

	return "unknown"
}

// GetIPListFromString returns list of ips in string form found in input string
// A return value of nil indicates no match
// FR-11.1: Now supports both IPv4 and IPv6 addresses using netip.ParseAddr
func GetIPListFromString(input string) []string {
	// First try to parse as URL to extract hostname
	u, err := url.Parse(input)
	if err == nil && u.Hostname() != "" {
		// Try to parse the hostname as an IP address
		if addr, err := netip.ParseAddr(u.Hostname()); err == nil {
			return []string{addr.String()}
		}
	}

	// Try to parse the input directly as an IP address (for bare IP strings)
	if addr, err := netip.ParseAddr(input); err == nil {
		return []string{addr.String()}
	}

	// Fallback to IPv4 regex for backward compatibility
	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	return re.FindAllString(input, -1)
}
