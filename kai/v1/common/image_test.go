// Copyright 2025 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func TestImageDigest(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	for _, tc := range []struct {
		name   string
		tag    string
		digest *string
		want   string
	}{
		{name: "unset", tag: "v1", want: "registry/example:v1"},
		{name: "empty", tag: "v1", digest: ptr.To(""), want: "registry/example:v1"},
		{name: "pinned", tag: "v1", digest: &digest, want: "registry/example:v1@" + digest},
		{name: "fips", tag: "v1-fips", digest: &digest, want: "registry/example:v1-fips@" + digest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			image := &Image{Repository: ptr.To("registry"), Name: ptr.To("example"), Tag: &tc.tag, Digest: tc.digest}
			image.SetDefaultsWhereNeeded()
			if got := image.Url(); got != tc.want {
				t.Errorf("Url() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestImageDigestDefaults(t *testing.T) {
	t.Setenv(DefaultRepositoryEnvVarName, "registry")
	t.Setenv(DefaultTagEnvVarName, "v1-fips")
	digest := "sha256:" + strings.Repeat("a", 64)
	image := &Image{Name: ptr.To("example"), Digest: ptr.To(digest)}
	image.SetDefaultsWhereNeeded()
	if got, want := image.Url(), "registry/example:v1-fips@"+digest; got != want {
		t.Errorf("Url() = %q, want %q", got, want)
	}
	other := &Image{Name: ptr.To("other")}
	other.SetDefaultsWhereNeeded()
	if got, want := other.Url(), "registry/other:v1-fips"; got != want {
		t.Errorf("Url() = %q, want %q", got, want)
	}
	copy := image.DeepCopy()
	*copy.Digest = "sha256:" + strings.Repeat("b", 64)
	if *image.Digest != digest {
		t.Fatal("DeepCopy shares the digest pointer")
	}
}

const (
	kaiDefaultRepository = "ghcr.io/kai-scheduler/kai-scheduler"
	defaultImageTag      = "latest"
)

func TestV1Alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V1Alpha1 Suite")
}

var _ = Describe("Image", func() {
	It("Set Defaults", func(ctx context.Context) {
		os.Setenv(DefaultRepositoryEnvVarName, kaiDefaultRepository)
		os.Setenv(DefaultTagEnvVarName, defaultImageTag)
		image := &Image{}
		image.SetDefaultsWhereNeeded()
		Expect(*image.Repository).To(Equal(kaiDefaultRepository))
		Expect(*image.PullPolicy).To(Equal(v1.PullIfNotPresent))
		Expect(*image.Tag).To(Equal(defaultImageTag))
	})

	It("Get URL", func(ctx context.Context) {
		image := &Image{
			Repository: ptr.To("ghcr.io/kai-scheduler/kai-scheduler"),
			Name:       ptr.To("example"),
			Tag:        ptr.To("test"),
		}
		Expect(image.Url()).To(Equal("ghcr.io/kai-scheduler/kai-scheduler/example:test"))
	})
})
