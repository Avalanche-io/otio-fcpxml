// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package fcpxml

import (
	"strings"
	"testing"

	"github.com/mrjoshuak/gotio/opentime"
)

func TestDecoder_SimpleProject(t *testing.T) {
	fcpxmlData := `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.9">
	<project name="Test Project">
		<sequence format="r1" duration="3600/24s">
			<spine>
				<video name="Clip 1" duration="1200/24s" start="0/24s"/>
				<gap name="Gap" duration="600/24s"/>
				<video name="Clip 2" duration="1800/24s" start="0/24s"/>
			</spine>
		</sequence>
	</project>
</fcpxml>`

	decoder := NewDecoder(strings.NewReader(fcpxmlData))
	timeline, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to decode FCPX XML: %v", err)
	}

	if timeline.Name() != "Test Project" {
		t.Errorf("Expected timeline name 'Test Project', got '%s'", timeline.Name())
	}

	// Check tracks
	videoTracks := timeline.VideoTracks()
	if len(videoTracks) != 1 {
		t.Errorf("Expected 1 video track, got %d", len(videoTracks))
	}

	// Check video track has 3 items (2 clips + 1 gap)
	if len(videoTracks) > 0 {
		children := videoTracks[0].Children()
		if len(children) != 3 {
			t.Errorf("Expected 3 children in video track, got %d", len(children))
		}
	}
}

func TestDecoder_WithMarkers(t *testing.T) {
	fcpxmlData := `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.9">
	<project name="Marker Test">
		<sequence format="r1">
			<spine>
				<video name="Clip with Marker" duration="2400/24s" start="0/24s">
					<marker start="600/24s" duration="0/24s" value="Test Marker" note="Marker comment"/>
				</video>
			</spine>
		</sequence>
	</project>
</fcpxml>`

	decoder := NewDecoder(strings.NewReader(fcpxmlData))
	timeline, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to decode FCPX XML: %v", err)
	}

	videoTracks := timeline.VideoTracks()
	if len(videoTracks) != 1 {
		t.Fatalf("Expected 1 video track, got %d", len(videoTracks))
	}

	children := videoTracks[0].Children()
	if len(children) != 1 {
		t.Fatalf("Expected 1 child in video track, got %d", len(children))
	}

	// Check markers
	clips := timeline.FindClips(nil, false)
	if len(clips) != 1 {
		t.Fatalf("Expected 1 clip, got %d", len(clips))
	}

	markers := clips[0].Markers()
	if len(markers) != 1 {
		t.Errorf("Expected 1 marker, got %d", len(markers))
	}

	if len(markers) > 0 {
		if markers[0].Name() != "Test Marker" {
			t.Errorf("Expected marker name 'Test Marker', got '%s'", markers[0].Name())
		}
		if markers[0].Comment() != "Marker comment" {
			t.Errorf("Expected marker comment 'Marker comment', got '%s'", markers[0].Comment())
		}
	}
}

func TestDecoder_ParseRationalTime(t *testing.T) {
	decoder := &Decoder{}

	tests := []struct {
		input    string
		expected opentime.RationalTime
		wantErr  bool
	}{
		{"1200/24s", opentime.NewRationalTime(1200, 24), false},
		{"0/24s", opentime.NewRationalTime(0, 24), false},
		{"3600/30s", opentime.NewRationalTime(3600, 30), false},
		{"1001/30000s", opentime.NewRationalTime(1001, 30000), false},
		{"", opentime.RationalTime{}, false},
		{"invalid", opentime.RationalTime{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := decoder.parseRationalTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRationalTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !result.StrictlyEqual(tt.expected) {
					t.Errorf("parseRationalTime(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestDecoder_LibraryStructure(t *testing.T) {
	fcpxmlData := `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.9">
	<library location="file:///Users/test/Library.fcpbundle">
		<event name="Test Event">
			<project name="Library Project">
				<sequence format="r1">
					<spine>
						<video name="Video in Library" duration="1200/24s" start="0/24s"/>
					</spine>
				</sequence>
			</project>
		</event>
	</library>
</fcpxml>`

	decoder := NewDecoder(strings.NewReader(fcpxmlData))
	timeline, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to decode FCPX XML with library: %v", err)
	}

	if timeline.Name() != "Library Project" {
		t.Errorf("Expected timeline name 'Library Project', got '%s'", timeline.Name())
	}
}

func TestDecoder_EmptyProject(t *testing.T) {
	fcpxmlData := `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.9">
	<project name="Empty Project">
		<sequence format="r1">
			<spine/>
		</sequence>
	</project>
</fcpxml>`

	decoder := NewDecoder(strings.NewReader(fcpxmlData))
	timeline, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to decode empty FCPX XML: %v", err)
	}

	if timeline.Name() != "Empty Project" {
		t.Errorf("Expected timeline name 'Empty Project', got '%s'", timeline.Name())
	}

	// Empty spine should result in empty tracks
	videoTracks := timeline.VideoTracks()
	audioTracks := timeline.AudioTracks()

	// Tracks may be created but should be empty
	totalChildren := 0
	for _, track := range videoTracks {
		totalChildren += len(track.Children())
	}
	for _, track := range audioTracks {
		totalChildren += len(track.Children())
	}

	if totalChildren != 0 {
		t.Errorf("Expected 0 children in empty project, got %d", totalChildren)
	}
}
