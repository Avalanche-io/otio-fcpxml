// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package fcpxml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio"
)

// Decoder reads FCPX XML and decodes it into an OTIO Timeline.
type Decoder struct {
	r io.Reader
}

// NewDecoder creates a new Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the FCPX XML document and converts it to an OTIO Timeline.
func (d *Decoder) Decode() (*gotio.Timeline, error) {
	var fcpxml FCPXML
	if err := xml.NewDecoder(d.r).Decode(&fcpxml); err != nil {
		return nil, fmt.Errorf("failed to parse FCPX XML: %w", err)
	}

	// Convert FCPX to OTIO Timeline
	return d.convertToTimeline(&fcpxml)
}

// convertToTimeline converts FCPXML to an OTIO Timeline.
func (d *Decoder) convertToTimeline(fcpxml *FCPXML) (*gotio.Timeline, error) {
	// Find the first project (either at root or in library/event)
	var project *Project
	if fcpxml.Project != nil {
		project = fcpxml.Project
	} else if fcpxml.Library != nil {
		if len(fcpxml.Library.Projects) > 0 {
			project = fcpxml.Library.Projects[0]
		} else if len(fcpxml.Library.Events) > 0 {
			for _, event := range fcpxml.Library.Events {
				if len(event.Projects) > 0 {
					project = event.Projects[0]
					break
				}
			}
		}
	} else if fcpxml.Event != nil && len(fcpxml.Event.Projects) > 0 {
		project = fcpxml.Event.Projects[0]
	}

	if project == nil {
		return nil, fmt.Errorf("no project found in FCPX XML")
	}

	// Create timeline
	timeline := gotio.NewTimeline(project.Name, nil, nil)

	// Convert sequence to tracks
	if project.Sequence != nil {
		if err := d.convertSequenceToTracks(project.Sequence, timeline); err != nil {
			return nil, err
		}
	}

	return timeline, nil
}

// convertSequenceToTracks converts a FCPX Sequence to OTIO tracks.
func (d *Decoder) convertSequenceToTracks(seq *Sequence, timeline *gotio.Timeline) error {
	if seq.Spine == nil {
		return nil
	}

	// FCPX uses a single spine, which we'll convert to separate video and audio tracks
	videoTrack := gotio.NewTrack("Video 1", nil, gotio.TrackKindVideo, nil, nil)
	audioTrack := gotio.NewTrack("Audio 1", nil, gotio.TrackKindAudio, nil, nil)

	// Process spine items
	for _, item := range seq.Spine.Items {
		switch v := item.(type) {
		case *Clip:
			// Convert clip to OTIO clips (may create both video and audio)
			if err := d.convertClip(v, videoTrack, audioTrack); err != nil {
				return err
			}
		case *Video:
			// Video-only clip
			if err := d.convertVideo(v, videoTrack); err != nil {
				return err
			}
		case *Audio:
			// Audio-only clip
			if err := d.convertAudio(v, audioTrack); err != nil {
				return err
			}
		case *Gap:
			// Gap/filler
			if err := d.convertGap(v, videoTrack, audioTrack); err != nil {
				return err
			}
		case *Transition:
			// Transitions are stored as metadata on adjacent clips
			// For now, skip direct processing
			continue
		case *RefClip:
			// Compound clip reference
			if err := d.convertRefClip(v, videoTrack, audioTrack); err != nil {
				return err
			}
		}
	}

	// Add tracks to timeline
	tracks := timeline.Tracks()
	if len(videoTrack.Children()) > 0 {
		tracks.AppendChild(videoTrack)
	}
	if len(audioTrack.Children()) > 0 {
		tracks.AppendChild(audioTrack)
	}

	return nil
}

// convertClip converts a FCPX Clip to OTIO clip(s).
func (d *Decoder) convertClip(clip *Clip, videoTrack, audioTrack *gotio.Track) error {
	// Parse duration
	duration, err := d.parseRationalTime(clip.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse clip duration: %w", err)
	}

	// Parse start time (source range start)
	var start opentime.RationalTime
	if clip.Start != "" {
		start, err = d.parseRationalTime(clip.Start)
		if err != nil {
			return fmt.Errorf("failed to parse clip start: %w", err)
		}
	}

	// Create source range
	sourceRange := opentime.NewTimeRange(start, duration)

	// Convert markers
	var markers []*gotio.Marker
	for _, m := range clip.Markers {
		marker, err := d.convertMarker(m)
		if err != nil {
			return err
		}
		markers = append(markers, marker)
	}

	// Create video clip if present
	if clip.Video != nil || clip.Ref != "" {
		ref := gotio.NewExternalReference("", "", nil, nil)
		otioClip := gotio.NewClip(clip.Name, ref, &sourceRange, nil, nil, markers, "", nil)
		videoTrack.AppendChild(otioClip)
	}

	// Create audio clip if present
	if clip.Audio != nil || clip.AudioDuration != "" {
		ref := gotio.NewExternalReference("", "", nil, nil)
		otioClip := gotio.NewClip(clip.Name, ref, &sourceRange, nil, nil, markers, "", nil)
		audioTrack.AppendChild(otioClip)
	}

	return nil
}

// convertVideo converts a FCPX Video element to OTIO clip.
func (d *Decoder) convertVideo(video *Video, videoTrack *gotio.Track) error {
	duration, err := d.parseRationalTime(video.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse video duration: %w", err)
	}

	var start opentime.RationalTime
	if video.Start != "" {
		start, err = d.parseRationalTime(video.Start)
		if err != nil {
			return fmt.Errorf("failed to parse video start: %w", err)
		}
	}

	sourceRange := opentime.NewTimeRange(start, duration)

	// Convert markers
	var markers []*gotio.Marker
	for _, m := range video.Markers {
		marker, err := d.convertMarker(m)
		if err != nil {
			return err
		}
		markers = append(markers, marker)
	}

	ref := gotio.NewExternalReference("", "", nil, nil)
	otioClip := gotio.NewClip(video.Name, ref, &sourceRange, nil, nil, markers, "", nil)
	videoTrack.AppendChild(otioClip)

	return nil
}

// convertAudio converts a FCPX Audio element to OTIO clip.
func (d *Decoder) convertAudio(audio *Audio, audioTrack *gotio.Track) error {
	duration, err := d.parseRationalTime(audio.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse audio duration: %w", err)
	}

	var start opentime.RationalTime
	if audio.Start != "" {
		start, err = d.parseRationalTime(audio.Start)
		if err != nil {
			return fmt.Errorf("failed to parse audio start: %w", err)
		}
	}

	sourceRange := opentime.NewTimeRange(start, duration)

	ref := gotio.NewExternalReference("", "", nil, nil)
	otioClip := gotio.NewClip(audio.Name, ref, &sourceRange, nil, nil, nil, "", nil)
	audioTrack.AppendChild(otioClip)

	return nil
}

// convertGap converts a FCPX Gap to OTIO gap(s).
func (d *Decoder) convertGap(gap *Gap, videoTrack, audioTrack *gotio.Track) error {
	duration, err := d.parseRationalTime(gap.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse gap duration: %w", err)
	}

	sourceRange := opentime.NewTimeRange(opentime.RationalTime{}, duration)

	// Add gap to both tracks
	videoGap := gotio.NewGap(gap.Name, &sourceRange, nil, nil, nil, nil)
	audioGap := gotio.NewGap(gap.Name, &sourceRange, nil, nil, nil, nil)

	videoTrack.AppendChild(videoGap)
	audioTrack.AppendChild(audioGap)

	return nil
}

// convertMarker converts a FCPX Marker to OTIO Marker.
func (d *Decoder) convertMarker(marker *Marker) (*gotio.Marker, error) {
	start, err := d.parseRationalTime(marker.Start)
	if err != nil {
		return nil, fmt.Errorf("failed to parse marker start: %w", err)
	}

	var duration opentime.RationalTime
	if marker.Duration != "" {
		duration, err = d.parseRationalTime(marker.Duration)
		if err != nil {
			return nil, fmt.Errorf("failed to parse marker duration: %w", err)
		}
	}

	markedRange := opentime.NewTimeRange(start, duration)

	// Use marker value as name, note as comment
	name := marker.Value
	comment := marker.Note

	return gotio.NewMarker(name, markedRange, gotio.MarkerColorGreen, comment, nil), nil
}

// convertRefClip converts a FCPX RefClip (compound clip) to OTIO Stack.
func (d *Decoder) convertRefClip(refClip *RefClip, videoTrack, audioTrack *gotio.Track) error {
	// Parse duration
	duration, err := d.parseRationalTime(refClip.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse ref-clip duration: %w", err)
	}

	// Parse start time
	var start opentime.RationalTime
	if refClip.Start != "" {
		start, err = d.parseRationalTime(refClip.Start)
		if err != nil {
			return fmt.Errorf("failed to parse ref-clip start: %w", err)
		}
	}

	// Create source range
	sourceRange := opentime.NewTimeRange(start, duration)

	// Convert markers
	var markers []*gotio.Marker
	for _, m := range refClip.Markers {
		marker, err := d.convertMarker(m)
		if err != nil {
			return err
		}
		markers = append(markers, marker)
	}

	// Create a Stack to represent the compound clip
	// In Python adapter, ref-clips become nested Stacks
	stack := gotio.NewStack(refClip.Name, &sourceRange, nil, nil, markers, nil)

	// Add metadata for compound clip reference
	metadata := map[string]interface{}{
		"fcpx_ref": refClip.Ref,
	}
	if refClip.SrcEnable != "" {
		metadata["fcpx_src_enable"] = refClip.SrcEnable
	}
	stack.SetMetadata(metadata)

	// Add to appropriate track based on srcEnable attribute
	if refClip.SrcEnable == "audio" {
		audioTrack.AppendChild(stack)
	} else {
		videoTrack.AppendChild(stack)
	}

	return nil
}

// parseRationalTime parses FCPX rational time format (e.g., "1001/30000s").
func (d *Decoder) parseRationalTime(s string) (opentime.RationalTime, error) {
	if s == "" {
		return opentime.RationalTime{}, nil
	}

	// Remove trailing 's' if present
	s = strings.TrimSuffix(s, "s")

	// Split on '/'
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		// Try parsing as a simple number (frames at default rate)
		value, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return opentime.RationalTime{}, fmt.Errorf("invalid rational time format: %s", s)
		}
		return opentime.NewRationalTime(value, 24), nil
	}

	// Parse numerator (value in frames)
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return opentime.RationalTime{}, fmt.Errorf("invalid rational time numerator: %s", parts[0])
	}

	// Parse denominator (rate)
	rate, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return opentime.RationalTime{}, fmt.Errorf("invalid rational time denominator: %s", parts[1])
	}

	// FCPX uses value/rate where value is in the rate's time base
	// Convert to OTIO format (value in frames, rate in fps)
	// The FCPX format is: frames/timebase, where timebase is the rate
	return opentime.NewRationalTime(value, rate), nil
}
