// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package fcpxml

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/Avalanche-io/gotio/opentime"
	"github.com/Avalanche-io/gotio/opentimelineio"
)

// Encoder writes an OTIO Timeline as FCPX XML.
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode converts an OTIO Timeline to FCPX XML and writes it to the output.
func (e *Encoder) Encode(timeline *opentimelineio.Timeline) error {
	// Convert OTIO Timeline to FCPXML
	fcpxml, err := e.convertFromTimeline(timeline)
	if err != nil {
		return err
	}

	// Write XML header
	if _, err := e.w.Write([]byte(xml.Header)); err != nil {
		return err
	}

	// Encode to XML
	encoder := xml.NewEncoder(e.w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(fcpxml); err != nil {
		return fmt.Errorf("failed to encode FCPX XML: %w", err)
	}

	// Add final newline
	if _, err := e.w.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

// convertFromTimeline converts an OTIO Timeline to FCPXML.
func (e *Encoder) convertFromTimeline(timeline *opentimelineio.Timeline) (*FCPXML, error) {
	// Create project
	project := &Project{
		Name: timeline.Name(),
	}

	// Create sequence from tracks
	sequence, err := e.convertTracksToSequence(timeline.Tracks())
	if err != nil {
		return nil, err
	}
	project.Sequence = sequence

	// Create FCPXML with the project
	fcpxml := &FCPXML{
		Version: "1.9",
		Project: project,
	}

	return fcpxml, nil
}

// convertTracksToSequence converts OTIO tracks to a FCPX Sequence.
func (e *Encoder) convertTracksToSequence(stack *opentimelineio.Stack) (*Sequence, error) {
	if stack == nil {
		return nil, fmt.Errorf("no tracks in timeline")
	}

	// Create spine
	spine := &Spine{
		Items: make([]interface{}, 0),
	}

	// Get video and audio tracks
	var videoItems []opentimelineio.Composable
	var audioItems []opentimelineio.Composable

	for _, child := range stack.Children() {
		if track, ok := child.(*opentimelineio.Track); ok {
			if track.Kind() == opentimelineio.TrackKindVideo {
				videoItems = append(videoItems, track.Children()...)
			} else if track.Kind() == opentimelineio.TrackKindAudio {
				audioItems = append(audioItems, track.Children()...)
			}
		}
	}

	// Convert items (prioritize video items for spine, add audio separately)
	for i, item := range videoItems {
		fcpItem, err := e.convertItem(item, true)
		if err != nil {
			return nil, err
		}
		if fcpItem != nil {
			spine.Items = append(spine.Items, fcpItem)
		}

		// Add corresponding audio if available
		if i < len(audioItems) {
			audioItem := audioItems[i]
			if audioClip, ok := audioItem.(*opentimelineio.Clip); ok {
				// Add audio to the clip
				if clip, ok := fcpItem.(*Clip); ok {
					audio, err := e.convertClipToAudio(audioClip)
					if err != nil {
						return nil, err
					}
					clip.Audio = audio
				}
			}
		}
	}

	// Create sequence
	sequence := &Sequence{
		Format:   "r1",
		Duration: "",
		Spine:    spine,
	}

	return sequence, nil
}

// convertItem converts an OTIO Composable to a FCPX Item.
func (e *Encoder) convertItem(item opentimelineio.Composable, isVideo bool) (Item, error) {
	switch v := item.(type) {
	case *opentimelineio.Clip:
		return e.convertClipToFCPX(v, isVideo)
	case *opentimelineio.Gap:
		return e.convertGapToFCPX(v)
	case *opentimelineio.Stack:
		return e.convertStackToRefClip(v)
	default:
		return nil, fmt.Errorf("unsupported item type: %T", item)
	}
}

// convertClipToFCPX converts an OTIO Clip to a FCPX Clip or Video.
func (e *Encoder) convertClipToFCPX(clip *opentimelineio.Clip, isVideo bool) (Item, error) {
	// Get duration
	duration, err := clip.Duration()
	if err != nil {
		return nil, fmt.Errorf("failed to get clip duration: %w", err)
	}

	// Get source range
	var start opentime.RationalTime
	if clip.SourceRange() != nil {
		start = clip.SourceRange().StartTime()
	}

	// Convert markers
	var markers []*Marker
	for _, m := range clip.Markers() {
		marker := e.convertMarkerToFCPX(m)
		markers = append(markers, marker)
	}

	if isVideo {
		// Create video clip
		video := &Video{
			Name:     clip.Name(),
			Duration: e.formatRationalTime(duration),
			Start:    e.formatRationalTime(start),
			Markers:  markers,
		}
		return video, nil
	}

	// Create audio clip
	audio := &Audio{
		Name:     clip.Name(),
		Duration: e.formatRationalTime(duration),
		Start:    e.formatRationalTime(start),
	}
	return audio, nil
}

// convertClipToAudio converts an OTIO Clip to a FCPX Audio element.
func (e *Encoder) convertClipToAudio(clip *opentimelineio.Clip) (*Audio, error) {
	duration, err := clip.Duration()
	if err != nil {
		return nil, fmt.Errorf("failed to get clip duration: %w", err)
	}

	var start opentime.RationalTime
	if clip.SourceRange() != nil {
		start = clip.SourceRange().StartTime()
	}

	audio := &Audio{
		Name:     clip.Name(),
		Duration: e.formatRationalTime(duration),
		Start:    e.formatRationalTime(start),
	}

	return audio, nil
}

// convertGapToFCPX converts an OTIO Gap to a FCPX Gap.
func (e *Encoder) convertGapToFCPX(gap *opentimelineio.Gap) (Item, error) {
	duration, err := gap.Duration()
	if err != nil {
		return nil, fmt.Errorf("failed to get gap duration: %w", err)
	}

	fcpGap := &Gap{
		Name:     gap.Name(),
		Duration: e.formatRationalTime(duration),
	}

	return fcpGap, nil
}

// convertMarkerToFCPX converts an OTIO Marker to a FCPX Marker.
func (e *Encoder) convertMarkerToFCPX(marker *opentimelineio.Marker) *Marker {
	markedRange := marker.MarkedRange()

	return &Marker{
		Start:    e.formatRationalTime(markedRange.StartTime()),
		Duration: e.formatRationalTime(markedRange.Duration()),
		Value:    marker.Name(),
		Note:     marker.Comment(),
	}
}

// convertStackToRefClip converts an OTIO Stack (compound clip) to a FCPX RefClip.
func (e *Encoder) convertStackToRefClip(stack *opentimelineio.Stack) (Item, error) {
	duration, err := stack.Duration()
	if err != nil {
		return nil, fmt.Errorf("failed to get stack duration: %w", err)
	}

	var start opentime.RationalTime
	if stack.SourceRange() != nil {
		start = stack.SourceRange().StartTime()
	}

	// Convert markers
	var markers []*Marker
	for _, m := range stack.Markers() {
		marker := e.convertMarkerToFCPX(m)
		markers = append(markers, marker)
	}

	// Get metadata for ref ID if available
	refID := "r1" // Default
	if metadata := stack.Metadata(); metadata != nil {
		if ref, ok := metadata["fcpx_ref"].(string); ok {
			refID = ref
		}
	}

	refClip := &RefClip{
		Name:     stack.Name(),
		Ref:      refID,
		Duration: e.formatRationalTime(duration),
		Start:    e.formatRationalTime(start),
		Markers:  markers,
	}

	return refClip, nil
}

// formatRationalTime converts an OTIO RationalTime to FCPX rational time format.
func (e *Encoder) formatRationalTime(rt opentime.RationalTime) string {
	if rt.Rate() <= 0 {
		return "0/1s"
	}
	// FCPX format: value/rate + "s"
	// Round value to avoid floating point precision issues
	value := int64(rt.Value())
	rate := int64(rt.Rate())
	return fmt.Sprintf("%d/%ds", value, rate)
}
