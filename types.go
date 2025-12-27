// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package fcpxml

import "encoding/xml"

// FCPXML represents the root element of a Final Cut Pro X XML document.
type FCPXML struct {
	XMLName   xml.Name   `xml:"fcpxml"`
	Version   string     `xml:"version,attr"`
	Resources *Resources `xml:"resources,omitempty"`
	Library   *Library   `xml:"library,omitempty"`
	Event     *Event     `xml:"event,omitempty"`
	Project   *Project   `xml:"project,omitempty"`
}

// Library represents a library element.
type Library struct {
	XMLName  xml.Name  `xml:"library"`
	Location string    `xml:"location,attr,omitempty"`
	Events   []*Event  `xml:"event,omitempty"`
	Projects []*Project `xml:"project,omitempty"`
}

// Event represents an event element.
type Event struct {
	XMLName  xml.Name    `xml:"event"`
	Name     string      `xml:"name,attr"`
	UID      string      `xml:"uid,attr,omitempty"`
	Projects []*Project  `xml:"project,omitempty"`
	Clips    []*Clip     `xml:"asset-clip,omitempty"`
	RefClips []*RefClip  `xml:"ref-clip,omitempty"`
}

// Project represents a project element.
type Project struct {
	XMLName  xml.Name  `xml:"project"`
	Name     string    `xml:"name,attr"`
	UID      string    `xml:"uid,attr,omitempty"`
	ModDate  string    `xml:"modDate,attr,omitempty"`
	Sequence *Sequence `xml:"sequence,omitempty"`
}

// Sequence represents a sequence element.
type Sequence struct {
	XMLName    xml.Name `xml:"sequence"`
	Format     string   `xml:"format,attr,omitempty"`
	Duration   string   `xml:"duration,attr,omitempty"`
	TCStart    string   `xml:"tcStart,attr,omitempty"`
	TCFormat   string   `xml:"tcFormat,attr,omitempty"`
	AudioLayout string  `xml:"audioLayout,attr,omitempty"`
	AudioRate  string   `xml:"audioRate,attr,omitempty"`
	Spine      *Spine   `xml:"spine,omitempty"`
}

// Spine represents the primary storyline/timeline.
type Spine struct {
	XMLName xml.Name `xml:"spine"`
	Items   []interface{}
}

// UnmarshalXML implements custom XML unmarshaling for Spine.
func (s *Spine) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	s.XMLName = start.Name
	s.Items = make([]interface{}, 0)

	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			var item interface{}
			switch t.Name.Local {
			case "asset-clip", "clip":
				item = &Clip{}
			case "video":
				item = &Video{}
			case "audio":
				item = &Audio{}
			case "gap":
				item = &Gap{}
			case "title":
				item = &Title{}
			case "transition":
				item = &Transition{}
			case "ref-clip":
				item = &RefClip{}
			default:
				// Skip unknown elements
				if err := d.Skip(); err != nil {
					return err
				}
				continue
			}

			if err := d.DecodeElement(item, &t); err != nil {
				return err
			}
			s.Items = append(s.Items, item)

		case xml.EndElement:
			if t == start.End() {
				return nil
			}
		}
	}
}

// Item is an interface for items that can appear in a spine or track.
type Item interface{}

// Clip represents a clip element (can be asset-clip, video, audio, etc).
type Clip struct {
	XMLName      xml.Name  `xml:"asset-clip"`
	Name         string    `xml:"name,attr,omitempty"`
	Ref          string    `xml:"ref,attr,omitempty"`
	Offset       string    `xml:"offset,attr,omitempty"`
	Start        string    `xml:"start,attr,omitempty"`
	Duration     string    `xml:"duration,attr,omitempty"`
	TCFormat     string    `xml:"tcFormat,attr,omitempty"`
	AudioStart   string    `xml:"audioStart,attr,omitempty"`
	AudioDuration string   `xml:"audioDuration,attr,omitempty"`
	AudioRole    string    `xml:"audioRole,attr,omitempty"`
	Markers      []*Marker `xml:"marker,omitempty"`
	Video        *Video    `xml:"video,omitempty"`
	Audio        *Audio    `xml:"audio,omitempty"`
}

// Video represents a video element.
type Video struct {
	XMLName  xml.Name  `xml:"video"`
	Name     string    `xml:"name,attr,omitempty"`
	Ref      string    `xml:"ref,attr,omitempty"`
	Offset   string    `xml:"offset,attr,omitempty"`
	Start    string    `xml:"start,attr,omitempty"`
	Duration string    `xml:"duration,attr,omitempty"`
	Markers  []*Marker `xml:"marker,omitempty"`
}

// Audio represents an audio element.
type Audio struct {
	XMLName  xml.Name   `xml:"audio"`
	Name     string     `xml:"name,attr,omitempty"`
	Ref      string     `xml:"ref,attr,omitempty"`
	Offset   string     `xml:"offset,attr,omitempty"`
	Start    string     `xml:"start,attr,omitempty"`
	Duration string     `xml:"duration,attr,omitempty"`
	Role     string     `xml:"role,attr,omitempty"`
	Channels []*Channel `xml:"audio-channel,omitempty"`
}

// Channel represents an audio channel element.
type Channel struct {
	XMLName  xml.Name `xml:"audio-channel"`
	Role     string   `xml:"role,attr,omitempty"`
	SrcCh    string   `xml:"srcCh,attr,omitempty"`
	Start    string   `xml:"start,attr,omitempty"`
	Duration string   `xml:"duration,attr,omitempty"`
}

// Gap represents a gap (filler) element.
type Gap struct {
	XMLName  xml.Name `xml:"gap"`
	Name     string   `xml:"name,attr,omitempty"`
	Offset   string   `xml:"offset,attr,omitempty"`
	Start    string   `xml:"start,attr,omitempty"`
	Duration string   `xml:"duration,attr,omitempty"`
}

// Marker represents a marker element.
type Marker struct {
	XMLName  xml.Name `xml:"marker"`
	Start    string   `xml:"start,attr,omitempty"`
	Duration string   `xml:"duration,attr,omitempty"`
	Value    string   `xml:"value,attr,omitempty"`
	Note     string   `xml:"note,attr,omitempty"`
	Completed bool    `xml:"completed,attr,omitempty"`
}

// Title represents a title element.
type Title struct {
	XMLName  xml.Name `xml:"title"`
	Name     string   `xml:"name,attr,omitempty"`
	Ref      string   `xml:"ref,attr,omitempty"`
	Offset   string   `xml:"offset,attr,omitempty"`
	Start    string   `xml:"start,attr,omitempty"`
	Duration string   `xml:"duration,attr,omitempty"`
}

// Transition represents a transition element.
type Transition struct {
	XMLName     xml.Name      `xml:"transition"`
	Name        string        `xml:"name,attr,omitempty"`
	Offset      string        `xml:"offset,attr,omitempty"`
	Duration    string        `xml:"duration,attr,omitempty"`
	FilterVideo *FilterVideo  `xml:"filter-video,omitempty"`
	FilterAudio *FilterAudio  `xml:"filter-audio,omitempty"`
}

// FilterVideo represents a filter-video element within a transition.
type FilterVideo struct {
	XMLName xml.Name `xml:"filter-video"`
	Ref     string   `xml:"ref,attr,omitempty"`
	Name    string   `xml:"name,attr,omitempty"`
	Params  []*Param `xml:"param,omitempty"`
}

// FilterAudio represents a filter-audio element within a transition.
type FilterAudio struct {
	XMLName xml.Name `xml:"filter-audio"`
	Ref     string   `xml:"ref,attr,omitempty"`
	Name    string   `xml:"name,attr,omitempty"`
	Params  []*Param `xml:"param,omitempty"`
}

// Param represents a parameter element within a filter.
type Param struct {
	XMLName xml.Name `xml:"param"`
	Name    string   `xml:"name,attr,omitempty"`
	Key     string   `xml:"key,attr,omitempty"`
	Value   string   `xml:"value,attr,omitempty"`
}

// Effect represents an effect element in resources.
type Effect struct {
	XMLName xml.Name `xml:"effect"`
	ID      string   `xml:"id,attr,omitempty"`
	Name    string   `xml:"name,attr,omitempty"`
	UID     string   `xml:"uid,attr,omitempty"`
}

// Media represents a media element (compound clip).
type Media struct {
	XMLName  xml.Name  `xml:"media"`
	ID       string    `xml:"id,attr,omitempty"`
	Name     string    `xml:"name,attr,omitempty"`
	UID      string    `xml:"uid,attr,omitempty"`
	ModDate  string    `xml:"modDate,attr,omitempty"`
	Sequence *Sequence `xml:"sequence,omitempty"`
}

// RefClip represents a ref-clip element (reference to compound clip).
type RefClip struct {
	XMLName         xml.Name  `xml:"ref-clip"`
	Name            string    `xml:"name,attr,omitempty"`
	Ref             string    `xml:"ref,attr,omitempty"`
	Offset          string    `xml:"offset,attr,omitempty"`
	Start           string    `xml:"start,attr,omitempty"`
	Duration        string    `xml:"duration,attr,omitempty"`
	SrcEnable       string    `xml:"srcEnable,attr,omitempty"`
	UseAudioSubroles bool     `xml:"useAudioSubroles,attr,omitempty"`
	Markers         []*Marker `xml:"marker,omitempty"`
}

// Keyword represents a keyword element.
type Keyword struct {
	XMLName  xml.Name `xml:"keyword"`
	Start    string   `xml:"start,attr,omitempty"`
	Duration string   `xml:"duration,attr,omitempty"`
	Value    string   `xml:"value,attr,omitempty"`
}

// Note represents a note element.
type Note struct {
	XMLName xml.Name `xml:"note"`
	Text    string   `xml:",chardata"`
}

// Metadata represents a metadata element.
type Metadata struct {
	XMLName xml.Name `xml:"metadata"`
	MD      []*MD    `xml:"md,omitempty"`
}

// MD represents a metadata key-value pair.
type MD struct {
	XMLName xml.Name `xml:"md"`
	Key     string   `xml:"key,attr,omitempty"`
	Value   string   `xml:"value,attr,omitempty"`
}

// Resources represents the resources element.
type Resources struct {
	XMLName xml.Name  `xml:"resources"`
	Formats []*Format `xml:"format,omitempty"`
	Assets  []*Asset  `xml:"asset,omitempty"`
	Media   []*Media  `xml:"media,omitempty"`
	Effects []*Effect `xml:"effect,omitempty"`
}

// Format represents a format element.
type Format struct {
	XMLName       xml.Name `xml:"format"`
	ID            string   `xml:"id,attr,omitempty"`
	Name          string   `xml:"name,attr,omitempty"`
	FrameDuration string   `xml:"frameDuration,attr,omitempty"`
	Width         string   `xml:"width,attr,omitempty"`
	Height        string   `xml:"height,attr,omitempty"`
	ColorSpace    string   `xml:"colorSpace,attr,omitempty"`
}

// Asset represents an asset element.
type Asset struct {
	XMLName       xml.Name `xml:"asset"`
	ID            string   `xml:"id,attr,omitempty"`
	Name          string   `xml:"name,attr,omitempty"`
	UID           string   `xml:"uid,attr,omitempty"`
	Src           string   `xml:"src,attr,omitempty"`
	Start         string   `xml:"start,attr,omitempty"`
	Duration      string   `xml:"duration,attr,omitempty"`
	Format        string   `xml:"format,attr,omitempty"`
	HasVideo      string   `xml:"hasVideo,attr,omitempty"`
	HasAudio      string   `xml:"hasAudio,attr,omitempty"`
	AudioSources  string   `xml:"audioSources,attr,omitempty"`
	AudioChannels string   `xml:"audioChannels,attr,omitempty"`
	AudioRate     string   `xml:"audioRate,attr,omitempty"`
}
