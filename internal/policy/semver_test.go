// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"testing"
)

func TestNewSemVer(t *testing.T) {
	cases := []struct {
		label        string
		semverRanges []string
		expectErr    bool
	}{
		{
			label:        "With valid range",
			semverRanges: []string{"1.0.x", "^1.0", "=1.0.0", "~1.0", ">=1.0", ">0,<2.0"},
		},
		{
			label:        "With invalid range",
			semverRanges: []string{"1.0.0p", "1x", "x1", "-1", "a", ""},
			expectErr:    true,
		},
	}

	for _, tt := range cases {
		for _, r := range tt.semverRanges {
			t.Run(tt.label, func(t *testing.T) {
				_, err := NewSemVer(r)
				if tt.expectErr && err == nil {
					t.Fatalf("expecting error, got nil for range value: '%s'", r)
				}
				if !tt.expectErr && err != nil {
					t.Fatalf("returned unexpected error: %s", err)
				}
			})
		}
	}
}

type testVersionedString struct {
	version string
}

func (v *testVersionedString) GetVersion() string {
	return v.version
}

func testVersionedStrings(versions ...string) []Versioned {
	//nolint:prealloc // Copied from another repo.
	var versioned []Versioned
	for _, v := range versions {
		versioned = append(versioned, &testVersionedString{v})
	}
	return versioned
}

//nolint:funlen // Long tests are ok.
func TestSemVer_Latest(t *testing.T) {
	cases := []struct {
		label           string
		semverRange     string
		versions        []Versioned
		expectedVersion string
		expectErr       bool
	}{
		{
			label: "With valid format",
			versions: testVersionedStrings(
				"1.0.0",
				"1.0.0.1",
				"1.0.0p",
				"1.0.1",
				"1.2.0",
				"0.1.0",
			),
			semverRange:     "1.0.x",
			expectedVersion: "1.0.1",
		},
		{
			label:           "With valid format prefix",
			versions:        testVersionedStrings("v1.2.3", "v1.0.0", "v0.1.0"),
			semverRange:     "1.0.x",
			expectedVersion: "v1.0.0",
		},
		{
			label:       "With invalid format prefix",
			versions:    testVersionedStrings("b1.2.3", "b1.0.0", "b0.1.0"),
			semverRange: "1.0.x",
			expectErr:   true,
		},
		{
			label:       "With empty list",
			versions:    testVersionedStrings(),
			semverRange: "1.0.x",
			expectErr:   true,
		},
		{
			label:       "With non-matching version list",
			versions:    testVersionedStrings("1.2.0"),
			semverRange: "1.0.x",
			expectErr:   true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			policy, err := NewSemVer(tt.semverRange)
			if err != nil {
				t.Fatalf("returned unexpected error: %s", err)
			}

			latest, err := policy.Latest(tt.versions)
			if tt.expectErr && err == nil {
				t.Fatalf("expecting error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("returned unexpected error: %s", err)
			}
			if latest == nil {
				if tt.expectedVersion != "" {
					t.Fatalf("expecting version, got nil")
				}
			} else {
				if latest.GetVersion() != tt.expectedVersion {
					t.Errorf(
						"incorrect computed version returned, got '%s', expected '%s'",
						latest,
						tt.expectedVersion,
					)
				}
			}
		})
	}
}
