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

package tracer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitTracing(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		prob    float64
		wantErr bool
	}{
		{
			name:    "valid URI and probability",
			uri:     "http://localhost:9411/api/v2/spans",
			prob:    0.5,
			wantErr: false,
		},
		{
			name:    "Negative probability",
			uri:     "http://localhost:9411/api/v2/spans",
			prob:    -1,
			wantErr: false,
		},
		{
			name:    "empty URI",
			uri:     "",
			prob:    0.5,
			wantErr: false,
		},
		{
			name:    "invalid UR",
			uri:     "invalid_uri",
			prob:    0.5,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InitTracing(tt.uri, tt.prob)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitTracing() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetTracer(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		span    string
		wantCtx context.Context
		wantErr bool
	}{
		{
			name:    "valid context and span",
			ctx:     context.Background(),
			span:    "test_span",
			wantCtx: context.Background(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCtx, span := GetTracer(tt.ctx, tt.span)
			assert.NotNil(t, gotCtx)
			assert.NotNil(t, span)
		})
	}
}
