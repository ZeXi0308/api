// Copyright 2025 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package v2alpha2

import (
	"context"
	"strings"
	"testing"

	"k8s.io/utils/ptr"
)

func leaf(name string, core bool) SubGroup {
	return SubGroup{Name: name, MinMember: ptr.To(int32(1)), Core: core}
}

func child(name, parent string, core bool) SubGroup {
	sg := leaf(name, core)
	sg.Parent = ptr.To(parent)
	return sg
}

func mid(name string, minSubGroup *int32, core bool) SubGroup {
	return SubGroup{Name: name, MinSubGroup: minSubGroup, Core: core}
}

func TestCoreMarksOnCreate(t *testing.T) {
	tests := []struct {
		name         string
		spec         PodGroupSpec
		wantErr      string
		wantWarnings int
	}{
		{
			name: "marks within root minSubGroup accepted",
			spec: PodGroupSpec{Preemptibility: SemiPreemptible, MinSubGroup: ptr.To(int32(1)),
				SubGroups: []SubGroup{leaf("a", true), leaf("b", false)}},
		},
		{
			name: "marks equal to root minSubGroup accepted",
			spec: PodGroupSpec{Preemptibility: SemiPreemptible, MinSubGroup: ptr.To(int32(2)),
				SubGroups: []SubGroup{leaf("a", true), leaf("b", true), leaf("c", false)}},
		},
		{
			name: "marks exceeding root minSubGroup rejected",
			spec: PodGroupSpec{Preemptibility: SemiPreemptible, MinSubGroup: ptr.To(int32(1)),
				SubGroups: []SubGroup{leaf("a", true), leaf("b", true)}},
			wantErr: "the PodGroup has 2 core child SubGroups, exceeding its effective minSubGroup (1)",
		},
		{
			name: "redundant marks with unset minSubGroup accepted",
			spec: PodGroupSpec{Preemptibility: SemiPreemptible,
				SubGroups: []SubGroup{leaf("a", true), leaf("b", true)}},
		},
		{
			name: "marks within nested minSubGroup accepted",
			spec: PodGroupSpec{Preemptibility: SemiPreemptible, MinSubGroup: ptr.To(int32(1)),
				SubGroups: []SubGroup{
					mid("p", ptr.To(int32(1)), true), child("x", "p", true), child("y", "p", false),
					leaf("z", false),
				}},
		},
		{
			name: "marks exceeding nested minSubGroup rejected",
			spec: PodGroupSpec{Preemptibility: SemiPreemptible, MinSubGroup: ptr.To(int32(1)),
				SubGroups: []SubGroup{
					mid("p", ptr.To(int32(1)), true), child("x", "p", true), child("y", "p", true),
				}},
			wantErr: `subgroup "p" has 2 core child SubGroups, exceeding its effective minSubGroup (1)`,
		},
		{
			name: "unset nested minSubGroup counts every child",
			spec: PodGroupSpec{Preemptibility: SemiPreemptible,
				SubGroups: []SubGroup{
					mid("p", nil, false), child("x", "p", true), child("y", "p", true),
				}},
		},
		{
			name: "core on preemptible warns and ignores over-quota marks",
			spec: PodGroupSpec{Preemptibility: Preemptible, MinSubGroup: ptr.To(int32(1)),
				SubGroups: []SubGroup{leaf("a", true), leaf("b", true)}},
			wantWarnings: 1,
		},
		{
			name: "core on non-preemptible warns",
			spec: PodGroupSpec{Preemptibility: NonPreemptible,
				SubGroups: []SubGroup{leaf("a", true)}},
			wantWarnings: 1,
		},
		{
			name: "no warning without marks on preemptible",
			spec: PodGroupSpec{Preemptibility: Preemptible,
				SubGroups: []SubGroup{leaf("a", false)}},
		},
		{
			name: "no warning on semi-preemptible",
			spec: PodGroupSpec{Preemptibility: SemiPreemptible,
				SubGroups: []SubGroup{leaf("a", true)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := (&PodGroup{}).ValidateCreate(context.Background(), &PodGroup{Spec: tt.spec})
			if tt.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("got error %v, want it to contain %q", err, tt.wantErr)
			}
			if len(warnings) != tt.wantWarnings {
				t.Fatalf("got warnings %v, want %d", warnings, tt.wantWarnings)
			}
		})
	}
}

func TestCoreMarksOnUpdate(t *testing.T) {
	semi := func(subGroups ...SubGroup) PodGroupSpec {
		return PodGroupSpec{Preemptibility: SemiPreemptible, MinSubGroup: ptr.To(int32(1)), SubGroups: subGroups}
	}
	nested := func(coreX, coreY bool, yParent string) PodGroupSpec {
		return semi(mid("p", ptr.To(int32(1)), false), mid("q", ptr.To(int32(1)), false),
			child("x", "p", coreX), child("y", yParent, coreY), child("pad", "q", false))
	}

	tests := []struct {
		name    string
		old     PodGroupSpec
		updated PodGroupSpec
		wantErr string
	}{
		{
			name:    "unchanged marks accepted",
			old:     semi(leaf("a", true), leaf("b", false)),
			updated: semi(leaf("a", true), leaf("b", false)),
		},
		{
			name:    "adding a mark rejected",
			old:     semi(leaf("a", false), leaf("b", false)),
			updated: semi(leaf("a", true), leaf("b", false)),
			wantErr: `cannot add core to subgroup "a"`,
		},
		{
			name:    "removing a mark rejected",
			old:     semi(leaf("a", true), leaf("b", false)),
			updated: semi(leaf("a", false), leaf("b", false)),
			wantErr: `cannot remove core from subgroup "a"`,
		},
		{
			name:    "moving a mark to a sibling rejected",
			old:     semi(leaf("a", true), leaf("b", false)),
			updated: semi(leaf("a", false), leaf("b", true)),
			wantErr: "cannot ",
		},
		{
			name:    "deleting a marked subgroup rejected",
			old:     semi(leaf("a", true), leaf("b", false)),
			updated: semi(leaf("b", false)),
			wantErr: `cannot remove core from subgroup "a"`,
		},
		{
			name:    "adding a new marked subgroup rejected",
			old:     semi(leaf("a", true), leaf("b", false)),
			updated: semi(leaf("a", true), leaf("b", false), leaf("c", true)),
			wantErr: `cannot add core to subgroup "c"`,
		},
		{
			name:    "adding a new unmarked subgroup accepted",
			old:     semi(leaf("a", true), leaf("b", false)),
			updated: semi(leaf("a", true), leaf("b", false), leaf("c", false)),
		},
		{
			name:    "nested marks unchanged accepted",
			old:     nested(true, false, "p"),
			updated: nested(true, false, "p"),
		},
		{
			name:    "nested mark added rejected",
			old:     nested(true, false, "p"),
			updated: nested(true, true, "p"),
			wantErr: `cannot add core to subgroup "y"`,
		},
		{
			name:    "nested mark reparented rejected",
			old:     nested(false, true, "p"),
			updated: nested(false, true, "q"),
			wantErr: `cannot move core subgroup "y"`,
		},
		{
			name: "marking preemptible group allowed (ignored)",
			old: PodGroupSpec{Preemptibility: Preemptible,
				SubGroups: []SubGroup{leaf("a", false)}},
			updated: PodGroupSpec{Preemptibility: Preemptible,
				SubGroups: []SubGroup{leaf("a", true)}},
		},
		{
			name: "unmarking while leaving semi-preemptible allowed",
			old:  semi(leaf("a", true), leaf("b", false)),
			updated: PodGroupSpec{Preemptibility: Preemptible, MinSubGroup: ptr.To(int32(1)),
				SubGroups: []SubGroup{leaf("a", false), leaf("b", false)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (&PodGroup{}).ValidateUpdate(context.Background(),
				&PodGroup{Spec: tt.old}, &PodGroup{Spec: tt.updated})
			if tt.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("got error %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
