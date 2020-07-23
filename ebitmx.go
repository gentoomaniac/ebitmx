package ebitmx

// https://www.onlinetool.io/xmltogo/

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
}

func (v Version) ToString() {
	fmt.Sprintf("%d.%d", v.Major, v.Minor)
}
func (v *Version) FromString(s string) error {
	fields := strings.Split(s, ".")
	if len(fields) != 2 {
		return errors.New("Version format wrong")
	}

	// error handling?
	v.Major, _ = strconv.Atoi(fields[0])
	v.Minor, _ = strconv.Atoi(fields[1])

	return nil
}

type TiledVersion struct {
	Major int
	Minor int
	Patch int
}

func (t TiledVersion) ToString() {
	fmt.Sprintf("%d.%d.%d", t.Major, t.Minor, t.Patch)
}
func (t *TiledVersion) FromString(s string) error {
	fields := strings.Split(s, ".")
	if len(fields) != 3 {
		return errors.New("Version format wrong")
	}

	// error handling?
	t.Major, _ = strconv.Atoi(fields[0])
	t.Minor, _ = strconv.Atoi(fields[1])
	t.Patch, _ = strconv.Atoi(fields[2])

	return nil
}

// Valid orientation types of a map
type Orientation string

const (
	Orthogonal Orientation = "orthogonal"
	Isometric              = "isometric"
	Staggered              = "staggered"
	hexagonal              = "hexagonal"
)

type RenderOrder string

const (
	RightDown RenderOrder = "right-down"
	RightUp               = "right-up"
	LeftDown              = "left-down"
	LeftUp                = "left-up"
)

type ObjectAlignment string

const (
	Unspecified ObjectAlignment = "unspecified"
	TopLeft                     = "topleft"
	Top                         = "top"
	TopRight                    = "topright"
	Left                        = "left"
	Center                      = "center"
	Right                       = "right"
	BottomLeft                  = "bottomleft"
	Bottom                      = "bottom"
	BottomRight                 = "bottomright"
)

type Tileset struct {
	firstgid        int
	source          string
	name            string
	tilewidth       int
	tileheight      int
	spacing         int
	margin          int
	tilecount       int
	columns         int
	objectalignment ObjectAlignment
}

type Tile struct {
}

type Layer struct {
	ID        int     `xml:"id"`
	Name      string  `xml:"name"`
	X         int     `xml:"x"`
	Y         int     `xml:"y"`
	Width     int     `xml:"width"`
	Height    int     `xml:"height"`
	Opacity   float64 `xml:"opacity"`
	Visible   bool
	Tintcolor string
	Offsetx   int
	Offsety   int
	tiles     []Tile
}
type ObjectLayer struct {
}
type ImageLayer struct {
}

type Map struct {
	version          Version
	layers           []Layer
	objectLayers     []ObjectLayer
	imageLayers      []ImageLayer
	orientation      Orientation
	renderOrder      RenderOrder
	compressionlevel int
	width            int
	height           int
	tilewidth        int
	tileheight       int
	// unsupported, placeholder
	hexsidelength int
	staggeraxis   int
	// #AARRGGBB
	backgroundcolor color.RGBA
	nextlayerid     int
	nextobjectid    int
	infinite        bool
}

func (m *Map) LoadFromFile(path string)
