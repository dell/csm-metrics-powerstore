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

package k8s_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/dell/csm-metrics-powerstore/internal/k8s/mocks"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_K8sPersistentVolumeFinder(t *testing.T) {
	type checkFn func(*testing.T, []k8s.VolumeInfo, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, _ []k8s.VolumeInfo, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput []k8s.VolumeInfo) func(t *testing.T, volumes []k8s.VolumeInfo, err error) {
		return func(t *testing.T, volumes []k8s.VolumeInfo, _ error) {
			assert.Equal(t, expectedOutput, volumes)
		}
	}

	hasError := func(t *testing.T, _ []k8s.VolumeInfo, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (k8s.VolumeFinder, []checkFn, *gomock.Controller){
		"success selecting the matching driver name with multiple volumes": func(*testing.T) (k8s.VolumeFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockVolumeGetter(ctrl)

			t1, err := time.Parse(time.RFC3339, "2020-07-28T20:00:00+00:00")
			assert.Nil(t, err)

			volumes := &corev1.PersistentVolumeList{
				Items: []corev1.PersistentVolume{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "persistent-volume-name",
							CreationTimestamp: metav1.Time{Time: t1},
						},
						Spec: corev1.PersistentVolumeSpec{
							Capacity: map[corev1.ResourceName]resource.Quantity{
								v1.ResourceStorage: resource.MustParse("16Gi"),
							},
							PersistentVolumeSource: corev1.PersistentVolumeSource{
								CSI: &corev1.CSIPersistentVolumeSource{
									Driver: "csi-powerstore.dellemc.com",
									VolumeAttributes: map[string]string{
										"arrayIP": "127.0.0.1",
									},
									VolumeHandle: "storage-system-volume-id/127.0.0.1/protocol",
								},
							},
							ClaimRef: &corev1.ObjectReference{
								Name:      "pvc-name",
								Namespace: "namespace-1",
								UID:       "pvc-uid",
							},
							StorageClassName: "storage-class-name",
						},
						Status: corev1.PersistentVolumeStatus{
							Phase: "Bound",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "persistent-volume-name",
							CreationTimestamp: metav1.Time{Time: t1},
						},
						Spec: corev1.PersistentVolumeSpec{
							Capacity: map[corev1.ResourceName]resource.Quantity{
								v1.ResourceStorage: resource.MustParse("16Gi"),
							},
							PersistentVolumeSource: corev1.PersistentVolumeSource{
								CSI: &corev1.CSIPersistentVolumeSource{
									Driver: "another-csi-driver.dellemc.com",
									VolumeAttributes: map[string]string{
										"arrayIP": "127.0.0.1",
									},
									VolumeHandle: "storage-system-volume-id/127.0.0.1/protocol",
								},
							},
							ClaimRef: &corev1.ObjectReference{
								Name:      "pvc-name",
								Namespace: "namespace-1",
								UID:       "pvc-uid",
							},
							StorageClassName: "storage-class-name",
						},
						Status: corev1.PersistentVolumeStatus{
							Phase: "Bound",
						},
					},
				},
			}

			api.EXPECT().GetPersistentVolumes().Times(1).Return(volumes, nil)

			finder := k8s.VolumeFinder{API: api, DriverNames: []string{"csi-powerstore.dellemc.com"}}
			return finder, check(hasNoError, checkExpectedOutput([]k8s.VolumeInfo{
				{
					Namespace:              "namespace-1",
					PersistentVolumeClaim:  "pvc-uid",
					PersistentVolumeStatus: "Bound",
					VolumeClaimName:        "pvc-name",
					PersistentVolume:       "persistent-volume-name",
					StorageClass:           "storage-class-name",
					Driver:                 "csi-powerstore.dellemc.com",
					ProvisionedSize:        "16Gi",
					VolumeHandle:           "storage-system-volume-id/127.0.0.1/protocol",
					CreatedTime:            t1.String(),
				},
			})), ctrl
		},
		"success selecting multiple volumes matching multiple driver names": func(*testing.T) (k8s.VolumeFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockVolumeGetter(ctrl)

			t1, err := time.Parse(time.RFC3339, "2020-07-28T20:00:00+00:00")
			assert.Nil(t, err)

			volumes := &corev1.PersistentVolumeList{
				Items: []corev1.PersistentVolume{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "persistent-volume-name",
							CreationTimestamp: metav1.Time{Time: t1},
						},
						Spec: corev1.PersistentVolumeSpec{
							Capacity: map[corev1.ResourceName]resource.Quantity{
								v1.ResourceStorage: resource.MustParse("16Gi"),
							},
							PersistentVolumeSource: corev1.PersistentVolumeSource{
								CSI: &corev1.CSIPersistentVolumeSource{
									Driver: "csi-powerstore.dellemc.com",
									VolumeAttributes: map[string]string{
										"arrayIP": "127.0.0.1",
									},
									VolumeHandle: "storage-system-volume-id/127.0.0.1/protocol",
								},
							},
							ClaimRef: &corev1.ObjectReference{
								Name:      "pvc-name",
								Namespace: "namespace-1",
								UID:       "pvc-uid",
							},
							StorageClassName: "storage-class-name",
						},
						Status: corev1.PersistentVolumeStatus{
							Phase: "Bound",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "persistent-volume-name-2",
							CreationTimestamp: metav1.Time{Time: t1},
						},
						Spec: corev1.PersistentVolumeSpec{
							Capacity: map[corev1.ResourceName]resource.Quantity{
								v1.ResourceStorage: resource.MustParse("8Gi"),
							},
							PersistentVolumeSource: corev1.PersistentVolumeSource{
								CSI: &corev1.CSIPersistentVolumeSource{
									Driver: "another-csi-driver.dellemc.com",
									VolumeAttributes: map[string]string{
										"arrayIP": "127.0.0.1",
									},
									VolumeHandle: "storage-system-volume-id/127.0.0.1/protocol",
								},
							},
							ClaimRef: &corev1.ObjectReference{
								Name:      "pvc-name-2",
								Namespace: "namespace-2",
								UID:       "pvc-uid-2",
							},
							StorageClassName: "storage-class-name-2",
						},
						Status: corev1.PersistentVolumeStatus{
							Phase: "Bound",
						},
					},
				},
			}

			api.EXPECT().GetPersistentVolumes().Times(1).Return(volumes, nil)

			finder := k8s.VolumeFinder{API: api, DriverNames: []string{"csi-powerstore.dellemc.com", "another-csi-driver.dellemc.com"}}
			return finder, check(hasNoError, checkExpectedOutput([]k8s.VolumeInfo{
				{
					Namespace:              "namespace-1",
					PersistentVolumeClaim:  "pvc-uid",
					PersistentVolumeStatus: "Bound",
					VolumeClaimName:        "pvc-name",
					PersistentVolume:       "persistent-volume-name",
					StorageClass:           "storage-class-name",
					Driver:                 "csi-powerstore.dellemc.com",
					ProvisionedSize:        "16Gi",
					VolumeHandle:           "storage-system-volume-id/127.0.0.1/protocol",
					CreatedTime:            t1.String(),
				},
				{
					Namespace:              "namespace-2",
					PersistentVolumeClaim:  "pvc-uid-2",
					PersistentVolumeStatus: "Bound",
					VolumeClaimName:        "pvc-name-2",
					PersistentVolume:       "persistent-volume-name-2",
					StorageClass:           "storage-class-name-2",
					Driver:                 "another-csi-driver.dellemc.com",
					ProvisionedSize:        "8Gi",
					VolumeHandle:           "storage-system-volume-id/127.0.0.1/protocol",
					CreatedTime:            t1.String(),
				},
			})), ctrl
		},
		"success selecting the matching driver name with non CSI Driver volumes": func(*testing.T) (k8s.VolumeFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockVolumeGetter(ctrl)

			t1, err := time.Parse(time.RFC3339, "2020-07-28T20:00:00+00:00")
			assert.Nil(t, err)

			volumes := &corev1.PersistentVolumeList{
				Items: []corev1.PersistentVolume{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "persistent-volume-name",
							CreationTimestamp: metav1.Time{Time: t1},
						},
						Spec: corev1.PersistentVolumeSpec{
							Capacity: map[corev1.ResourceName]resource.Quantity{
								v1.ResourceStorage: resource.MustParse("16Gi"),
							},
							PersistentVolumeSource: corev1.PersistentVolumeSource{
								CSI: &corev1.CSIPersistentVolumeSource{
									Driver: "csi-powerstore.dellemc.com",
									VolumeAttributes: map[string]string{
										"arrayIP": "127.0.0.1",
									},
									VolumeHandle: "storage-system-volume-id/127.0.0.1/protocol",
								},
							},
							ClaimRef: &corev1.ObjectReference{
								Name:      "pvc-name",
								Namespace: "namespace-1",
								UID:       "pvc-uid",
							},
							StorageClassName: "storage-class-name",
						},
						Status: corev1.PersistentVolumeStatus{
							Phase: "Bound",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "persistent-volume-name",
							CreationTimestamp: metav1.Time{Time: t1},
						},
						Spec: corev1.PersistentVolumeSpec{
							Capacity: map[corev1.ResourceName]resource.Quantity{
								v1.ResourceStorage: resource.MustParse("16Gi"),
							},
							PersistentVolumeSource: corev1.PersistentVolumeSource{},
							ClaimRef: &corev1.ObjectReference{
								Name:      "pvc-name",
								Namespace: "namespace-1",
								UID:       "pvc-uid",
							},
							StorageClassName: "storage-class-name",
						},
						Status: corev1.PersistentVolumeStatus{
							Phase: "Bound",
						},
					},
				},
			}

			api.EXPECT().GetPersistentVolumes().Times(1).Return(volumes, nil)

			finder := k8s.VolumeFinder{API: api, DriverNames: []string{"csi-powerstore.dellemc.com"}, Logger: logrus.New()}
			return finder, check(hasNoError, checkExpectedOutput([]k8s.VolumeInfo{
				{
					Namespace:              "namespace-1",
					PersistentVolumeClaim:  "pvc-uid",
					PersistentVolumeStatus: "Bound",
					VolumeClaimName:        "pvc-name",
					PersistentVolume:       "persistent-volume-name",
					StorageClass:           "storage-class-name",
					Driver:                 "csi-powerstore.dellemc.com",
					ProvisionedSize:        "16Gi",
					VolumeHandle:           "storage-system-volume-id/127.0.0.1/protocol",
					CreatedTime:            t1.String(),
				},
			})), ctrl
		},
		"success filtering the persistent volumes which do not have claims": func(*testing.T) (k8s.VolumeFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockVolumeGetter(ctrl)

			t1, err := time.Parse(time.RFC3339, "2020-07-28T20:00:00+00:00")
			assert.Nil(t, err)

			volumes := &corev1.PersistentVolumeList{
				Items: []corev1.PersistentVolume{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "persistent-volume-name",
							CreationTimestamp: metav1.Time{Time: t1},
						},
						Spec: corev1.PersistentVolumeSpec{
							Capacity: map[corev1.ResourceName]resource.Quantity{
								v1.ResourceStorage: resource.MustParse("16Gi"),
							},
							PersistentVolumeSource: corev1.PersistentVolumeSource{
								CSI: &corev1.CSIPersistentVolumeSource{
									Driver: "csi-powerstore.dellemc.com",
									VolumeAttributes: map[string]string{
										"arrayIP": "127.0.0.1",
									},
									VolumeHandle: "storage-system-volume-id/127.0.0.1/protocol",
								},
							},
							ClaimRef: &corev1.ObjectReference{
								Name:      "pvc-name",
								Namespace: "namespace-1",
								UID:       "pvc-uid",
							},
							StorageClassName: "storage-class-name",
						},
						Status: corev1.PersistentVolumeStatus{
							Phase: "Bound",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "persistent-volume-name2",
							CreationTimestamp: metav1.Time{Time: t1},
						},
						Spec: corev1.PersistentVolumeSpec{
							Capacity: map[corev1.ResourceName]resource.Quantity{
								v1.ResourceStorage: resource.MustParse("16Gi"),
							},
							PersistentVolumeSource: corev1.PersistentVolumeSource{
								CSI: &corev1.CSIPersistentVolumeSource{
									Driver: "csi-powerstore.dellemc.com",
									VolumeAttributes: map[string]string{
										"arrayIP": "127.0.0.1",
									},
									VolumeHandle: "storage-system-volume-id/127.0.0.1/protocol",
								},
							},
							ClaimRef:         nil,
							StorageClassName: "storage-class-name",
						},
						Status: corev1.PersistentVolumeStatus{
							Phase: "Available",
						},
					},
				},
			}

			api.EXPECT().GetPersistentVolumes().Times(1).Return(volumes, nil)

			finder := k8s.VolumeFinder{API: api, DriverNames: []string{"csi-powerstore.dellemc.com"}, Logger: logrus.New()}
			return finder, check(hasNoError, checkExpectedOutput([]k8s.VolumeInfo{
				{
					Namespace:              "namespace-1",
					PersistentVolumeClaim:  "pvc-uid",
					PersistentVolumeStatus: "Bound",
					VolumeClaimName:        "pvc-name",
					PersistentVolume:       "persistent-volume-name",
					StorageClass:           "storage-class-name",
					Driver:                 "csi-powerstore.dellemc.com",
					ProvisionedSize:        "16Gi",
					VolumeHandle:           "storage-system-volume-id/127.0.0.1/protocol",
					CreatedTime:            t1.String(),
				},
			})), ctrl
		},
		"error calling k8s": func(*testing.T) (k8s.VolumeFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockVolumeGetter(ctrl)
			api.EXPECT().GetPersistentVolumes().Times(1).Return(nil, errors.New("error"))
			finder := k8s.VolumeFinder{API: api}
			return finder, check(hasError), ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			finder, checkFns, ctrl := tc(t)
			volumes, err := finder.GetPersistentVolumes(context.Background())
			for _, checkFn := range checkFns {
				checkFn(t, volumes, err)
			}
			ctrl.Finish()
		})
	}
}
