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

package service_test

import (
	"context"
	"testing"

	"github.com/dell/csm-metrics-powerstore/internal/service"
	"go.opentelemetry.io/otel"
)

func TestMetricsWrapper_Record(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}
	volumeMetas := []interface{}{
		&service.VolumeMeta{
			ID: "123",
		},
	}
	spaceMetas := []interface{}{
		&service.SpaceVolumeMeta{
			ID: "123",
		},
	}
	type args struct {
		ctx              context.Context
		meta             interface{}
		readBW           float32
		writeBW          float32
		readIOPS         float32
		writeIOPS        float32
		readLatency      float32
		writeLatency     float32
		syncronizationBW float32
		mirrorBW         float32
		dataRemaining    float32
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:              context.Background(),
				meta:             volumeMetas[0],
				readBW:           1,
				writeBW:          2,
				readIOPS:         3,
				writeIOPS:        4,
				readLatency:      5,
				writeLatency:     6,
				syncronizationBW: 7,
				mirrorBW:         8,
				dataRemaining:    9,
			},
			wantErr: false,
		},
		{
			name: "fail",
			mw:   mw,
			args: args{
				ctx:              context.Background(),
				meta:             spaceMetas[0],
				readBW:           1,
				writeBW:          2,
				readIOPS:         3,
				writeIOPS:        4,
				readLatency:      5,
				writeLatency:     6,
				syncronizationBW: 7,
				mirrorBW:         8,
				dataRemaining:    9,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.Record(tt.args.ctx, tt.args.meta, tt.args.readBW, tt.args.writeBW, tt.args.readIOPS, tt.args.writeIOPS, tt.args.readLatency, tt.args.writeLatency, tt.args.syncronizationBW, tt.args.mirrorBW, tt.args.dataRemaining); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.Record() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_RecordSpaceMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}
	spaceMetas := []interface{}{
		&service.SpaceVolumeMeta{
			ID: "123",
		},
	}
	volumeMetas := []interface{}{
		&service.VolumeMeta{
			ID: "123",
		},
	}
	type args struct {
		ctx                context.Context
		meta               interface{}
		logicalProvisioned int64
		logicalUsed        int64
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:                context.Background(),
				meta:               spaceMetas[0],
				logicalProvisioned: 1,
				logicalUsed:        2,
			},
			wantErr: false,
		},
		{
			name: "fail",
			mw:   mw,
			args: args{
				ctx:                context.Background(),
				meta:               volumeMetas[0],
				logicalProvisioned: 1,
				logicalUsed:        2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordSpaceMetrics(tt.args.ctx, tt.args.meta, tt.args.logicalProvisioned, tt.args.logicalUsed); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordSpaceMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_RecordArraySpaceMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}

	type args struct {
		ctx                context.Context
		arrayID            string
		driver             string
		logicalProvisioned int64
		logicalUsed        int64
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:                context.Background(),
				arrayID:            "123",
				driver:             "driver",
				logicalProvisioned: 1,
				logicalUsed:        2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordArraySpaceMetrics(tt.args.ctx, tt.args.arrayID, tt.args.driver, tt.args.logicalProvisioned, tt.args.logicalUsed); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordArraySpaceMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_RecordStorageClassSpaceMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}
	type args struct {
		ctx                context.Context
		storageclass       string
		driver             string
		logicalProvisioned int64
		logicalUsed        int64
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:                context.Background(),
				storageclass:       "storageclass",
				driver:             "driver",
				logicalProvisioned: 1,
				logicalUsed:        2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordStorageClassSpaceMetrics(tt.args.ctx, tt.args.storageclass, tt.args.driver, tt.args.logicalProvisioned, tt.args.logicalUsed); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordStorageClassSpaceMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_RecordFileSystemMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}
	volumeMetas := []interface{}{
		&service.VolumeMeta{
			ID: "123",
		},
	}
	type args struct {
		ctx           context.Context
		meta          interface{}
		readBW        float32
		writeBW       float32
		readIOPS      float32
		writeIOPS     float32
		readLatency   float32
		writeLatency  float32
		syncBW        float32
		mirrorBW      float32
		dataRemaining float32
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:           context.Background(),
				meta:          volumeMetas[0],
				readBW:        1,
				writeBW:       2,
				readIOPS:      3,
				writeIOPS:     4,
				readLatency:   5,
				writeLatency:  6,
				syncBW:        7,
				mirrorBW:      8,
				dataRemaining: 9,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordFileSystemMetrics(tt.args.ctx, tt.args.meta, tt.args.readBW, tt.args.writeBW, tt.args.readIOPS, tt.args.writeIOPS, tt.args.readLatency, tt.args.writeLatency, tt.args.syncBW, tt.args.mirrorBW, tt.args.dataRemaining); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordFileSystemMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
