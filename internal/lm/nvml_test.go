package lm

import (
	"testing"

	"github.com/stretchr/testify/require"

	spec "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
	"github.com/NVIDIA/k8s-device-plugin/internal/resource"
	rt "github.com/NVIDIA/k8s-device-plugin/internal/resource/testing"
)

func TestMigCapabilityLabeler(t *testing.T) {
	testCases := []struct {
		description    string
		devices        []resource.Device
		expectedError  bool
		expectedLabels map[string]string
	}{
		{
			description: "no devices returns empty labels",
		},
		{
			description: "single non-mig capable device returns mig.capable as false",
			devices: []resource.Device{
				rt.NewFullGPU(),
			},
			expectedLabels: map[string]string{
				"nvidia.com/mig.capable": "false",
			},
		},
		{
			description: "multiple non-mig capable devices returns mig.capable as false",
			devices: []resource.Device{
				rt.NewFullGPU(),
				rt.NewFullGPU(),
			},
			expectedLabels: map[string]string{
				"nvidia.com/mig.capable": "false",
			},
		},
		{
			description: "single mig capable device returns mig.capable as true",
			devices: []resource.Device{
				rt.NewMigEnabledDevice(),
			},
			expectedLabels: map[string]string{
				"nvidia.com/mig.capable": "true",
			},
		},
		{
			description: "one mig capable device among multiple returns mig.capable as true",
			devices: []resource.Device{
				rt.NewFullGPU(),
				rt.NewMigEnabledDevice(),
			},
			expectedLabels: map[string]string{
				"nvidia.com/mig.capable": "true",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			nvmlMock := rt.NewManagerMockWithDevices(tc.devices...)

			migCapabilityLabeler, _ := newMigCapabilityLabeler(nvmlMock)

			labels, err := migCapabilityLabeler.Labels()
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.EqualValues(t, tc.expectedLabels, labels)
		})
	}
}

func TestSharingLabeler(t *testing.T) {
	testCases := []struct {
		descrition     string
		config         *spec.Config
		expectedLabels map[string]string
	}{
		{
			descrition: "nil config",
			expectedLabels: map[string]string{
				"nvidia.com/sharing.mps.enabled": "false",
			},
		},
		{
			descrition: "empty config",
			config:     &spec.Config{},
			expectedLabels: map[string]string{
				"nvidia.com/sharing.mps.enabled": "false",
			},
		},
		{
			descrition: "config with timeslicing replicas",
			config: &spec.Config{
				Sharing: spec.Sharing{
					TimeSlicing: spec.ReplicatedResources{
						Resources: []spec.ReplicatedResource{
							{
								Replicas: 2,
							},
						},
					},
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/sharing.mps.enabled": "false",
			},
		},
		{
			descrition: "config with no mps replicas",
			config: &spec.Config{
				Sharing: spec.Sharing{
					MPS: &spec.ReplicatedResources{
						Resources: []spec.ReplicatedResource{
							{
								Replicas: 1,
							},
						},
					},
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/sharing.mps.enabled": "false",
			},
		},
		{
			descrition: "config with mps replicas",
			config: &spec.Config{
				Sharing: spec.Sharing{
					MPS: &spec.ReplicatedResources{
						Resources: []spec.ReplicatedResource{
							{
								Replicas: 2,
							},
						},
					},
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/sharing.mps.enabled": "true",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.descrition, func(t *testing.T) {
			require.EqualValues(t, tc.expectedLabels, newSharingLabeler(tc.config))
		})
	}
}
