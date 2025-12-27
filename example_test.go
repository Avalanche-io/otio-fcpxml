// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package fcpxml_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mrjoshuak/gotio/opentime"
	"github.com/mrjoshuak/gotio/opentimelineio"
	"github.com/mrjoshuak/gotio/otio-fcpxml"
)

func ExampleDecoder() {
	fcpxmlData := `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.9">
	<project name="Example Project">
		<sequence format="r1">
			<spine>
				<video name="Shot 1" duration="1200/24s" start="0/24s"/>
				<video name="Shot 2" duration="1800/24s" start="0/24s"/>
			</spine>
		</sequence>
	</project>
</fcpxml>`

	decoder := fcpxml.NewDecoder(strings.NewReader(fcpxmlData))
	timeline, err := decoder.Decode()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Timeline: %s\n", timeline.Name())
	fmt.Printf("Video Tracks: %d\n", len(timeline.VideoTracks()))

	// Output:
	// Timeline: Example Project
	// Video Tracks: 1
}

func ExampleEncoder() {
	// Create a simple timeline
	timeline := opentimelineio.NewTimeline("Example Timeline", nil, nil)

	// Create a video track
	videoTrack := opentimelineio.NewTrack("Video 1", nil, opentimelineio.TrackKindVideo, nil, nil)

	// Add some clips with source ranges (required for encoding)
	ref1 := opentimelineio.NewExternalReference("", "", nil, nil)
	sourceRange1 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(1200, 24),
	)
	clip1 := opentimelineio.NewClip("Clip 1", ref1, &sourceRange1, nil, nil, nil, "", nil)

	ref2 := opentimelineio.NewExternalReference("", "", nil, nil)
	sourceRange2 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(1800, 24),
	)
	clip2 := opentimelineio.NewClip("Clip 2", ref2, &sourceRange2, nil, nil, nil, "", nil)

	videoTrack.AppendChild(clip1)
	videoTrack.AppendChild(clip2)

	// Add track to timeline
	timeline.Tracks().AppendChild(videoTrack)

	// Encode to FCPX XML
	var buf bytes.Buffer
	encoder := fcpxml.NewEncoder(&buf)
	if err := encoder.Encode(timeline); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// The output would be valid FCPX XML
	fmt.Println("FCPX XML generated successfully")

	// Output:
	// FCPX XML generated successfully
}
