// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package k8s_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/dell/csm-metrics-powerstore/internal/k8s"

	"k8s.io/client-go/kubernetes"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func Test_GetPersistentVolumes(t *testing.T) {
	type checkFn func(*testing.T, *corev1.PersistentVolumeList, error)
	type connectFn func(*k8s.API) error
	type configFn func() (*rest.Config, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, volumes *corev1.PersistentVolumeList, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput *corev1.PersistentVolumeList) func(t *testing.T, volumes *corev1.PersistentVolumeList, err error) {
		return func(t *testing.T, volumes *corev1.PersistentVolumeList, err error) {
			assert.Equal(t, expectedOutput, volumes)
		}
	}

	hasError := func(t *testing.T, volumes *corev1.PersistentVolumeList, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (connectFn, configFn, []checkFn){
		"success": func(*testing.T) (connectFn, configFn, []checkFn) {

			volumes := &corev1.PersistentVolumeList{
				Items: []corev1.PersistentVolume{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "persistent-volume-name",
						},
					},
				},
			}
			connect := func(api *k8s.API) error {
				api.Client = fake.NewSimpleClientset(volumes)
				return nil
			}
			return connect, nil, check(hasNoError, checkExpectedOutput(volumes))
		},
		"error connecting": func(*testing.T) (connectFn, configFn, []checkFn) {
			connect := func(api *k8s.API) error {
				return errors.New("error")
			}
			return connect, nil, check(hasError)
		},
		"error getting a valid config": func(*testing.T) (connectFn, configFn, []checkFn) {
			inClusterConfig := func() (*rest.Config, error) {
				return nil, errors.New("error")
			}
			return nil, inClusterConfig, check(hasError)
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			connectFn, inClusterConfig, checkFns := tc(t)
			k8sclient := &k8s.API{}
			if connectFn != nil {
				oldConnectFn := k8s.ConnectFn
				defer func() { k8s.ConnectFn = oldConnectFn }()
				k8s.ConnectFn = connectFn
			}
			if inClusterConfig != nil {
				oldInClusterConfig := k8s.InClusterConfigFn
				defer func() { k8s.InClusterConfigFn = oldInClusterConfig }()
				k8s.InClusterConfigFn = inClusterConfig
			}
			volumes, err := k8sclient.GetPersistentVolumes()
			for _, checkFn := range checkFns {
				checkFn(t, volumes, err)
			}
		})
	}

}

func Test_InClusterConfigFn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, err := k8s.InClusterConfigFn()
		assert.Error(t, err)
	})
}

func Test_NewForConfigError(t *testing.T) {
	k8sapi := &k8s.API{}

	oldInClusterConfigFn := k8s.InClusterConfigFn
	defer func() { k8s.InClusterConfigFn = oldInClusterConfigFn }()
	k8s.InClusterConfigFn = func() (*rest.Config, error) {
		return new(rest.Config), nil
	}

	oldNewConfigFn := k8s.NewConfigFn
	defer func() { k8s.NewConfigFn = oldNewConfigFn }()
	expected := "could not create Clientset from KubeConfig"
	k8s.NewConfigFn = func(config *rest.Config) (*kubernetes.Clientset, error) {
		return nil, fmt.Errorf(expected)
	}

	_, err := k8sapi.GetPersistentVolumes()
	assert.True(t, err != nil)
	if err != nil {
		assert.Equal(t, expected, err.Error())
	}
}
