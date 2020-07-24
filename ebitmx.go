package ebitmx

// https://www.onlinetool.io/xmltogo/

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	"io/ioutil"
	"strings"
)

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
	Text            string          `xml:",chardata"`
	FirstGid        uint32          `xml:"firstgid,attr"`
	Source          string          `xml:"source,attr"`
	Name            string          `xml:"name,attr"`
	Tilewidth       int             `xml:"tilewidth,attr"`
	Tileheight      int             `xml:"tileheight,attr"`
	Spacing         int             `xml:"spacing,attr"`
	Margin          int             `xml:"margin,attr"`
	Tilecount       int             `xml:"tilecount,attr"`
	Columns         int             `xml:"colums,attr"`
	Objectalignment ObjectAlignment `xml:"objectalignment,attr"`
}

const (
	FLIPPED_HORIZONTALLY_FLAG uint32 = 0x80000000
	FLIPPED_VERTICALLY_FLAG   uint32 = 0x40000000
	FLIPPED_DIAGONALLY_FLAG   uint32 = 0x20000000
)

type Tile struct {
	GlobalTileID        uint32
	FlippedHorizontally bool
	FlippedVertically   bool
	FlippedDiagonally   bool
	Tileset             *Tileset
	Position            image.Point
}

func TileFromByteArray(data []byte) *Tile {
	t := &Tile{}
	encodedID := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24

	t.FlippedHorizontally = (encodedID & FLIPPED_HORIZONTALLY_FLAG) > 1
	t.FlippedVertically = (encodedID & FLIPPED_VERTICALLY_FLAG) > 1
	t.FlippedDiagonally = (encodedID & FLIPPED_DIAGONALLY_FLAG) > 1

	t.GlobalTileID = encodedID & ((FLIPPED_DIAGONALLY_FLAG | FLIPPED_HORIZONTALLY_FLAG | FLIPPED_VERTICALLY_FLAG) ^ 0xffffffff)

	return t
}

type DataEncoding string

const (
	Base64 DataEncoding = "base64"
	CSV                 = "csv"
)

type Compression string

const (
	Gzip Compression = "gzip"
	Zlib             = "zlib"
	Zstd             = "zstd"
)

type Layer struct {
	Text      string  `xml:",chardata"`
	ID        uint    `xml:"id,attr"`
	Name      string  `xml:"name,attr"`
	X         int     `xml:"x,attr"`
	Y         int     `xml:"y,attr"`
	Width     int     `xml:"width,attr"`
	Height    int     `xml:"height,attr"`
	Opacity   float64 `xml:"opacity,attr"`
	Visible   bool    `xml:"visible,attr"`
	Tintcolor string  `xml:"tintcolor,attr"`
	Offsetx   int     `xml:"offsetx,attr"`
	Offsety   int     `xml:"offsety,attr"`
	Tiles     []*Tile
	Data      struct {
		Text        string       `xml:",chardata"`
		Encoding    DataEncoding `xml:"encoding,attr"`
		Compression Compression  `xml:"compression,attr"`
	} `xml:"data"`
}

func (l *Layer) DecodeData(gameMap *TmxMap) error {
	if l.Data.Encoding == Base64 {
		byteArray, error := base64.StdEncoding.DecodeString(strings.TrimSpace(l.Data.Text))
		if error != nil {
			fmt.Printf("Error decoding data: %s", error)
			return error
		}

		tileNum := 0
		for i := 0; i < len(byteArray)-4; i += 4 {
			newTile := TileFromByteArray(byteArray[i : i+4])
			for _, tileset := range gameMap.Tilesets {
				if newTile.GlobalTileID >= tileset.FirstGid {
					newTile.Tileset = &tileset
				}
			}
			tileNum++
			newTile.Position = image.Point{tileNum % gameMap.Width, tileNum / gameMap.Height}
			l.Tiles = append(l.Tiles, newTile)
		}
	}
	return nil
}

type TmxMap struct {
	XMLName          xml.Name    `xml:"map"`
	Text             string      `xml:",chardata"`
	Version          string      `xml:"version,attr"`
	Tiledversion     string      `xml:"tiledversion,attr"`
	Orientation      Orientation `xml:"orientation,attr"`
	Renderorder      RenderOrder `xml:"renderorder,attr"`
	Compressionlevel int         `xml:"compressionlevel,attr"`
	Width            int         `xml:"width,attr"`
	Height           int         `xml:"height,attr"`
	Tilewidth        int         `xml:"tilewidth,attr"`
	Tileheight       int         `xml:"tileheight,attr"`
	Hexsidelength    int         `xml:"hexsidelength,attr"`
	Staggeraxis      int         `xml:"staggeraxis,attr"`
	Backgroundcolor  string      `xml:"backgroundcolor,attr"`
	Infinite         int         `xml:"infinite,attr"`
	Nextlayerid      int         `xml:"nextlayerid,attr"`
	Nextobjectid     int         `xml:"nextobjectid,attr"`
	Tilesets         []Tileset   `xml:"tileset"`
	Layers           []Layer     `xml:"layer"`
}

func LoadFromFile(path string) (*TmxMap, error) {
	gameMap := &TmxMap{}

	data, error := ioutil.ReadFile(path)
	if error != nil {
		return gameMap, error
	}

	_ = xml.Unmarshal([]byte(data), &gameMap)

	for _, l := range gameMap.Layers {
		l.DecodeData(gameMap)
	}

	return gameMap, error
}
