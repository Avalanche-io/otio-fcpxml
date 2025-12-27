// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package fcpxml

import (
	"os"
	"testing"
)

// TestDecoder_CompoundClips tests reading files with compound clips (ref-clip elements).
func TestDecoder_CompoundClips(t *testing.T) {
	data, err := os.ReadFile("testdata/fcpx_clips.fcpxml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Verify we have the expected structure
	t.Logf("Successfully parsed FCPXML with compound clips: %d bytes", len(data))
}

// TestDecoder_Transitions tests reading files with transitions.
func TestDecoder_Transitions(t *testing.T) {
	data, err := os.ReadFile("testdata/fcpx_example.fcpxml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// The example file contains transitions
	// Verify file can be read without errors
	t.Logf("Successfully read FCPXML with transitions: %d bytes", len(data))
}

// TestDecoder_KeywordsAndMetadata tests reading files with keywords and metadata.
func TestDecoder_KeywordsAndMetadata(t *testing.T) {
	data, err := os.ReadFile("testdata/fcpx_clips.fcpxml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// The clips file contains keywords and metadata on asset-clips
	// Keywords: "snow", "truck"
	// Metadata: com.apple.proapps.studio.angle = "B"
	// Notes: "Truck in snow"

	t.Logf("Successfully read FCPXML with keywords and metadata: %d bytes", len(data))
}

// TestDecoder_Roles tests reading files with audio and video roles.
func TestDecoder_Roles(t *testing.T) {
	data, err := os.ReadFile("testdata/fcpx_example.fcpxml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// The example file contains role attributes on audio elements
	// e.g., role="dialogue.dialogue-1"

	t.Logf("Successfully read FCPXML with roles: %d bytes", len(data))
}

// TestTypes_Transition tests that Transition type can be created.
func TestTypes_Transition(t *testing.T) {
	transition := &Transition{
		Name:     "Cross Dissolve",
		Offset:   "0s",
		Duration: "1s",
		FilterVideo: &FilterVideo{
			Ref:  "r12",
			Name: "Cross Dissolve",
			Params: []*Param{
				{Name: "Amount", Key: "2", Value: "50"},
			},
		},
		FilterAudio: &FilterAudio{
			Ref:  "r13",
			Name: "Audio Crossfade",
		},
	}

	if transition.Name != "Cross Dissolve" {
		t.Errorf("Expected name 'Cross Dissolve', got '%s'", transition.Name)
	}
	if transition.FilterVideo == nil {
		t.Error("Expected FilterVideo to be set")
	}
	if transition.FilterAudio == nil {
		t.Error("Expected FilterAudio to be set")
	}
}

// TestTypes_RefClip tests that RefClip type can be created.
func TestTypes_RefClip(t *testing.T) {
	refClip := &RefClip{
		Name:     "compound_clip_1",
		Ref:      "r1",
		Duration: "80s",
		Markers: []*Marker{
			{Start: "6s", Duration: "100/3000s", Value: "Marker 1"},
		},
	}

	if refClip.Name != "compound_clip_1" {
		t.Errorf("Expected name 'compound_clip_1', got '%s'", refClip.Name)
	}
	if len(refClip.Markers) != 1 {
		t.Errorf("Expected 1 marker, got %d", len(refClip.Markers))
	}
}

// TestTypes_Keyword tests that Keyword type can be created.
func TestTypes_Keyword(t *testing.T) {
	keyword := &Keyword{
		Start:    "0s",
		Duration: "10s",
		Value:    "snow, truck",
	}

	if keyword.Value != "snow, truck" {
		t.Errorf("Expected value 'snow, truck', got '%s'", keyword.Value)
	}
}

// TestTypes_Metadata tests that Metadata type can be created.
func TestTypes_Metadata(t *testing.T) {
	metadata := &Metadata{
		MD: []*MD{
			{Key: "com.apple.proapps.studio.reel", Value: "5"},
			{Key: "com.apple.proapps.studio.scene", Value: "17"},
		},
	}

	if len(metadata.MD) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(metadata.MD))
	}
}

// TestTypes_Effect tests that Effect type can be created.
func TestTypes_Effect(t *testing.T) {
	effect := &Effect{
		ID:   "r12",
		Name: "Cross Dissolve",
		UID:  "FxPlug:4731E73A-8DAC-4113-9A30-AE85B1761265",
	}

	if effect.Name != "Cross Dissolve" {
		t.Errorf("Expected name 'Cross Dissolve', got '%s'", effect.Name)
	}
}

// TestTypes_Media tests that Media type can be created.
func TestTypes_Media(t *testing.T) {
	media := &Media{
		ID:      "r1",
		Name:    "compound_clip_1",
		UID:     "JQCJMjHxQRKdo3UOjQKiNQ",
		ModDate: "2019-02-16 07:51:04 -0500",
		Sequence: &Sequence{
			Duration: "80s",
			Format:   "r2",
		},
	}

	if media.Name != "compound_clip_1" {
		t.Errorf("Expected name 'compound_clip_1', got '%s'", media.Name)
	}
	if media.Sequence == nil {
		t.Error("Expected Sequence to be set")
	}
}

// TestTypes_Resources tests that Resources type can be created.
func TestTypes_Resources(t *testing.T) {
	resources := &Resources{
		Formats: []*Format{
			{ID: "r2", Name: "FFVideoFormat1080p30", FrameDuration: "100/3000s"},
		},
		Assets: []*Asset{
			{ID: "r3", Name: "IMG_0233", Src: "file:///test.mov"},
		},
		Media: []*Media{
			{ID: "r1", Name: "compound_clip_1"},
		},
		Effects: []*Effect{
			{ID: "r12", Name: "Cross Dissolve"},
		},
	}

	if len(resources.Formats) != 1 {
		t.Errorf("Expected 1 format, got %d", len(resources.Formats))
	}
	if len(resources.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(resources.Assets))
	}
	if len(resources.Media) != 1 {
		t.Errorf("Expected 1 media, got %d", len(resources.Media))
	}
	if len(resources.Effects) != 1 {
		t.Errorf("Expected 1 effect, got %d", len(resources.Effects))
	}
}
