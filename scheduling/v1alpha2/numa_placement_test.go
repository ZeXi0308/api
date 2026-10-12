// Copyright 2026 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestBindRequestMemoryGroupsSerialization(t *testing.T) {
	t.Run("legacy records remain compatible", func(t *testing.T) {
		legacy := `{"podName":"pod","selectedNode":"node","predictedNUMAZones":[{"zone":"node-0","amount":{"memory":"64Gi"}}]}`
		var spec BindRequestSpec
		require.NoError(t, json.Unmarshal([]byte(legacy), &spec))
		require.Nil(t, spec.PredictedNUMAMemoryGroups)
		encoded, err := json.Marshal(spec)
		require.NoError(t, err)
		require.JSONEq(t, legacy, string(encoded))
	})

	t.Run("optional groups preserve the annotation format", func(t *testing.T) {
		spec := BindRequestSpec{
			PodName:      "pod",
			SelectedNode: "node",
			PredictedNUMAMemoryGroups: []NUMAMemoryGroupPlacement{{
				MemoryNodes: []string{"node-0", "node-1"},
				Amount: v1.ResourceList{
					v1.ResourceMemory: resource.MustParse("120Gi"),
					"hugepages-2Mi":   resource.MustParse("4Mi"),
				},
			}},
		}
		encoded, err := json.Marshal(spec)
		require.NoError(t, err)
		require.JSONEq(t, `{"podName":"pod","selectedNode":"node","predictedNUMAMemoryGroups":[{"memoryNodes":["node-0","node-1"],"amount":{"memory":"120Gi","hugepages-2Mi":"4Mi"}}]}`, string(encoded))
		var decoded BindRequestSpec
		require.NoError(t, json.Unmarshal(encoded, &decoded))
		require.Equal(t, spec, decoded)
		var legacyReader struct {
			PodName      string `json:"podName"`
			SelectedNode string `json:"selectedNode"`
		}
		require.NoError(t, json.Unmarshal(encoded, &legacyReader))
		require.Equal(t, spec.PodName, legacyReader.PodName)
		require.Equal(t, spec.SelectedNode, legacyReader.SelectedNode)
	})

	t.Run("nil and empty groups are omitted", func(t *testing.T) {
		for _, groups := range [][]NUMAMemoryGroupPlacement{nil, {}} {
			encoded, err := json.Marshal(BindRequestSpec{PodName: "pod", PredictedNUMAMemoryGroups: groups})
			require.NoError(t, err)
			require.JSONEq(t, `{"podName":"pod"}`, string(encoded))
		}
	})
}

func TestNUMAMemoryGroupDeepCopy(t *testing.T) {
	amount := resource.MustParse("120Gi")
	amount.ToDec()
	original := &NUMAMemoryGroupPlacement{
		MemoryNodes: []string{"node-0", "node-1"},
		Amount:      v1.ResourceList{v1.ResourceMemory: amount},
	}
	clone := original.DeepCopy()
	clone.MemoryNodes[0] = "node-9"
	cloneAmount := clone.Amount[v1.ResourceMemory]
	cloneAmount.Add(resource.MustParse("1Gi"))
	clone.Amount[v1.ResourceMemory] = cloneAmount
	clone.Amount["hugepages-2Mi"] = resource.MustParse("2Mi")
	require.Equal(t, []string{"node-0", "node-1"}, original.MemoryNodes)
	originalAmount := original.Amount[v1.ResourceMemory]
	require.Zero(t, originalAmount.Cmp(resource.MustParse("120Gi")))
	require.Len(t, original.Amount, 1)
	require.Nil(t, (&NUMAMemoryGroupPlacement{}).DeepCopy().MemoryNodes)
	require.Nil(t, (&NUMAMemoryGroupPlacement{}).DeepCopy().Amount)
	var absent *NUMAMemoryGroupPlacement
	require.Nil(t, absent.DeepCopy())
}

func TestBindRequestMemoryGroupsDeepCopy(t *testing.T) {
	original := &BindRequest{Spec: BindRequestSpec{PredictedNUMAMemoryGroups: []NUMAMemoryGroupPlacement{{
		MemoryNodes: []string{"node-0", "node-1"},
		Amount:      v1.ResourceList{v1.ResourceMemory: resource.MustParse("120Gi")},
	}}}}
	clone := original.DeepCopy()
	clone.Spec.PredictedNUMAMemoryGroups[0].MemoryNodes[0] = "node-9"
	clone.Spec.PredictedNUMAMemoryGroups[0].Amount[v1.ResourceMemory] = resource.MustParse("1Gi")
	clone.Spec.PredictedNUMAMemoryGroups = append(clone.Spec.PredictedNUMAMemoryGroups, NUMAMemoryGroupPlacement{})
	require.Len(t, original.Spec.PredictedNUMAMemoryGroups, 1)
	require.Equal(t, []string{"node-0", "node-1"}, original.Spec.PredictedNUMAMemoryGroups[0].MemoryNodes)
	amount := original.Spec.PredictedNUMAMemoryGroups[0].Amount[v1.ResourceMemory]
	require.Zero(t, amount.Cmp(resource.MustParse("120Gi")))
	require.Nil(t, (&BindRequest{}).DeepCopy().Spec.PredictedNUMAMemoryGroups)
	empty := (&BindRequest{Spec: BindRequestSpec{PredictedNUMAMemoryGroups: []NUMAMemoryGroupPlacement{}}}).DeepCopy()
	require.NotNil(t, empty.Spec.PredictedNUMAMemoryGroups)
	require.Empty(t, empty.Spec.PredictedNUMAMemoryGroups)
}
