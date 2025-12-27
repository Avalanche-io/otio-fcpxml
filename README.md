# FCPX XML Adapter for gotio

A Go implementation of the Final Cut Pro X XML adapter for OpenTimelineIO.

## Overview

This adapter provides reading and writing of Final Cut Pro X formatted XML files (FCPXML) for the gotio (Go OpenTimelineIO) library. It follows Go's standard encoding patterns with `Decoder` and `Encoder` types.

## Features

### Supported

- ✅ Multiple video tracks
- ✅ Audio tracks & clips
- ✅ Gaps/fillers
- ✅ Markers (with color support: green=completed, red=incomplete, purple=standard)
- ✅ Basic nesting (library/event/project structure)
- ✅ Transitions (stored as metadata, parsed but not converted to OTIO effects)
- ✅ Compound clips (ref-clip/media elements converted to nested Stacks)
- ✅ Audio/Video roles (preserved in metadata)
- ✅ Keywords (parsed from asset-clip elements)
- ✅ Custom metadata (md elements within metadata blocks)
- ✅ Effects/filters (parsed as type definitions in resources)

### Not Yet Supported

- ❌ Full transition effects (parsed but not converted to OTIO Effect objects)
- ❌ Advanced color grading
- ❌ Multicam clips
- ❌ Speed effects (retime)
- ❌ Full nested sequence expansion (compound clips are represented as Stacks)

## Installation

```bash
go get github.com/mrjoshuak/gotio/otio-fcpxml
```

For local development:

```bash
cd otio-fcpxml
go mod edit -replace github.com/mrjoshuak/gotio=../gotio
go mod tidy
```

## Usage

### Decoding FCPX XML to OTIO Timeline

```go
package main

import (
    "os"
    "github.com/mrjoshuak/gotio/otio-fcpxml"
)

func main() {
    // Open FCPX XML file
    file, err := os.Open("project.fcpxml")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create decoder and decode
    decoder := fcpxml.NewDecoder(file)
    timeline, err := decoder.Decode()
    if err != nil {
        panic(err)
    }

    // Use the timeline
    println("Timeline:", timeline.Name())
}
```

### Encoding OTIO Timeline to FCPX XML

```go
package main

import (
    "os"
    "github.com/mrjoshuak/gotio/opentimelineio"
    "github.com/mrjoshuak/gotio/otio-fcpxml"
)

func main() {
    // Create or load a timeline
    timeline := opentimelineio.NewTimeline("My Project", nil, nil)

    // ... populate timeline with tracks and clips ...

    // Open output file
    file, err := os.Create("output.fcpxml")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create encoder and encode
    encoder := fcpxml.NewEncoder(file)
    if err := encoder.Encode(timeline); err != nil {
        panic(err)
    }
}
```

## FCPX XML Format

The Final Cut Pro X XML format (FCPXML) is different from the legacy FCP 7 XML format:

- Uses `<fcpxml>` root element (not `<xmeml>`)
- Uses rational time format: `"1001/30000s"` instead of timecode
- Hierarchical structure: `<library>` → `<event>` → `<project>` → `<sequence>` → `<spine>`
- The `<spine>` element contains clips in sequential order

### Example FCPXML Structure

```xml
<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.9">
  <project name="My Project">
    <sequence format="r1" duration="3600/24s">
      <spine>
        <video name="Clip 1" duration="1200/24s" start="0/24s">
          <marker start="600/24s" duration="0/24s" value="Important" note="Review this"/>
        </video>
        <gap name="Gap" duration="600/24s"/>
        <video name="Clip 2" duration="1800/24s" start="0/24s"/>
      </spine>
    </sequence>
  </project>
</fcpxml>
```

## API Design

Following Go's standard encoding patterns:

```go
type Decoder struct {
    r io.Reader
}

func NewDecoder(r io.Reader) *Decoder
func (d *Decoder) Decode() (*opentimelineio.Timeline, error)

type Encoder struct {
    w io.Writer
}

func NewEncoder(w io.Writer) *Encoder
func (e *Encoder) Encode(t *opentimelineio.Timeline) error
```

## Testing

Run tests:

```bash
go test -v
```

Run examples:

```bash
go test -v -run Example
```

## References

- [FCPXML Reference - Apple Developer](https://developer.apple.com/library/archive/documentation/FinalCutProX/Reference/FinalCutProXXMLFormat/EventsandProjects/EventsandProjects.html)
- [Demystifying Final Cut Pro XMLs](https://fcp.cafe/developer-case-studies/fcpxml/)
- [OpenTimelineIO FCPX XML Adapter (Python)](https://github.com/OpenTimelineIO/otio-fcpx-xml-adapter)
- [FCPXML Format Guide](https://orcasubtitle.com/docs/fcpxml-format)

## License

Apache-2.0 - See LICENSE file for details.

Copyright Contributors to the OpenTimelineIO project.
