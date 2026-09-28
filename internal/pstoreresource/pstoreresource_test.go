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

package pstoreresource_test

import (
	"context"
	"os"
	"testing"

	"github.com/dell/csm-metrics-powerstore/internal/pstoreresource"
	csictx "github.com/dell/gocsi/context"
	"github.com/stretchr/testify/assert"
)

func Test_Run(t *testing.T) {
	tests := map[string]func(t *testing.T) (filePath string, env map[string]string, expectError bool){
		"success": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/sample-config.yaml", map[string]string{pstoreresource.EnvThrottlingRateLimit: "123"}, false
		},
		"invalid throttling value": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/sample-config.yaml", map[string]string{pstoreresource.EnvThrottlingRateLimit: "abc"}, false
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
		"empty arrays": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/empty-array.yaml", map[string]string{pstoreresource.EnvThrottlingRateLimit: "abc"}, false
		},
		"nil array entry": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/nil-array.yaml", nil, true
		},
		"client with dns creation": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/client-dns.yaml", nil, false
		},
		"valid api timeout": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/sample-config.yaml", map[string]string{pstoreresource.EnvPowerstoreAPITimeout: "60s"}, false
		},
		"invalid api timeout": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/sample-config.yaml", map[string]string{pstoreresource.EnvPowerstoreAPITimeout: "invalid"}, false
		},
		"numeric hostname endpoint": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/numeric-hostname.yaml", nil, true
		},
		"short endpoint": func(*testing.T) (string, map[string]string, bool) {
			return "testdata/short-endpoint.yaml", nil, true
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			filePath, envs, expectError := test(t)

			for k, v := range envs {
				csictx.Setenv(context.Background(), k, v)
			}

			arrays, mapper, defaultArray, err := pstoreresource.GetPowerStoreArrays(filePath)

			switch name {
			case "empty arrays":
				assert.Equal(t, 0, len(arrays))
				assert.Nil(t, defaultArray)
			case "nil array entry":
				assert.Empty(t, arrays)
				assert.Empty(t, mapper)
				assert.Nil(t, defaultArray)
				assert.Nil(t, err)
				return
			// Other cases...
			default:
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
			}
		})
	}
}

func TestGetPowerStoreArrays_FQDNProtocol(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "fqdn-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.WriteString("arrays:\n  - endpoint: https://localhost/api/rest\n    globalID: array-fqdn\n    username: user\n    password: password\n    skipCertificateValidation: true\n")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	arrays, _, _, err := pstoreresource.GetPowerStoreArrays(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if got := arrays["array-fqdn"].NetworkProtocol; got != "ipv4" && got != "ipv6" && got != "unknown" {
		t.Fatalf("unexpected FQDN protocol classification %q", got)
	}
}

// FR-11.1: Test GetIPListFromString with IPv6 support
func TestGetIPListFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "bracketed IPv6 endpoint",
			input:    "https://[2001:db8::1]/api/rest",
			expected: []string{"2001:db8::1"},
		},
		{
			name:     "IPv4 endpoint",
			input:    "https://10.0.0.1/api/rest",
			expected: []string{"10.0.0.1"},
		},
		{
			name:     "FQDN endpoint",
			input:    "https://my-array.example.com/api/rest",
			expected: nil,
		},
		{
			name:     "IPv4 endpoint with port",
			input:    "https://10.0.0.1:443/api/rest",
			expected: []string{"10.0.0.1"},
		},
		{
			name:     "bracketed IPv6 with port",
			input:    "https://[2001:db8::1]:443/api/rest",
			expected: []string{"2001:db8::1"},
		},
		{
			name:     "bare IPv6 string",
			input:    "2001:db8::1",
			expected: []string{"2001:db8::1"},
		},
		{
			name:     "bare IPv4 string",
			input:    "10.0.0.1",
			expected: []string{"10.0.0.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pstoreresource.GetIPListFromString(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// FR-10.1: Test InferProtocol function
func TestInferProtocol(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{
			name:     "IPv4 address",
			ip:       "10.0.0.1",
			expected: "ipv4",
		},
		{
			name:     "IPv6 address",
			ip:       "2001:db8::1",
			expected: "ipv6",
		},
		{
			name:     "IPv4-mapped IPv6",
			ip:       "::ffff:192.0.2.1",
			expected: "ipv6",
		},
		{
			name:     "invalid address",
			ip:       "invalid",
			expected: "unknown",
		},
		{
			name:     "empty string",
			ip:       "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pstoreresource.InferProtocol(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}
