package ebitmx

// https://www.onlinetool.io/xmltogo/

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
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

type TSXFile struct {
	XMLName      xml.Name `xml:"tileset"`
	Text         string   `xml:",chardata"`
	Version      string   `xml:"version,attr"`
	TiledVersion string   `xml:"tiledversion,attr"`
	Name         string   `xml:"name,attr"`
	TileWidth    int      `xml:"tilewidth,attr"`
	TileHeight   int      `xml:"tileheight,attr"`
	TileCount    int      `xml:"tilecount,attr"`
	Columns      int      `xml:"columns,attr"`
	Image        struct {
		Text   string `xml:",chardata"`
		Source string `xml:"source,attr"`
		Width  int    `xml:"width,attr"`
		Height int    `xml:"height,attr"`
	} `xml:"image"`
}

type Tileset struct {
	Text               string          `xml:",chardata"`
	FirstGid           uint32          `xml:"firstgid,attr"`
	Source             string          `xml:"source,attr"`
	Name               string          `xml:"name,attr"`
	TileWidth          int             `xml:"tilewidth,attr"`
	TileHeight         int             `xml:"tileheight,attr"`
	Spacing            int             `xml:"spacing,attr"`
	Margin             int             `xml:"margin,attr"`
	TileCount          int             `xml:"tilecount,attr"`
	Columns            int             `xml:"colums,attr"`
	Objectalignment    ObjectAlignment `xml:"objectalignment,attr"`
	TilesetEbitenImage *ebiten.Image
	TilesetImage       image.Image
	Version            string `xml:"version,attr"`
	Tiledversion       string `xml:"tiledversion,attr"`
	Tiles              map[int]*ebiten.Image
}

func (t *Tileset) LoadFromTsx(path string) error {
	tsxFile := &TSXFile{}
	absTSXPath, error := filepath.Abs(filepath.Join(path, t.Source))
	if error != nil {
		return error
	}

	data, error := ioutil.ReadFile(absTSXPath)
	if error != nil {
		return error
	}
	_ = xml.Unmarshal([]byte(data), &tsxFile)

	t.Version = tsxFile.Version
	t.Tiledversion = tsxFile.TiledVersion
	t.TileWidth = tsxFile.TileWidth
	t.TileHeight = tsxFile.TileHeight
	t.TileCount = tsxFile.TileCount
	t.Columns = tsxFile.Columns

	absImgPath, error := filepath.Abs(filepath.Join(filepath.Dir(absTSXPath), tsxFile.Image.Source))
	if error != nil {
	}

	t.TilesetEbitenImage, t.TilesetImage, error = ebitenutil.NewImageFromFile(absImgPath, ebiten.FilterDefault)
	if error != nil {
		return fmt.Errorf("Failed loading texture: %s\n", error)
	}

	fmt.Printf("Pre-loading all tiles from '%s'...\n", t.Name)
	t.Tiles = make(map[int]*ebiten.Image)
	tileNum := 0
	for ; tileNum < t.TileCount; tileNum++ {
		x0 := (tileNum % t.Columns) * t.TileWidth
		y0 := (tileNum / t.Columns) * t.TileWidth

		tileRectangle := image.Rect(x0, y0, x0+t.TileWidth, y0+t.TileHeight)
		t.Tiles[tileNum] = t.TilesetEbitenImage.SubImage(tileRectangle).(*ebiten.Image)
	}
	fmt.Printf("%d tiles loaded.\n", tileNum)

	return nil
}

const (
	FLIPPED_HORIZONTALLY_FLAG uint32 = 0x80000000
	FLIPPED_VERTICALLY_FLAG   uint32 = 0x40000000
	FLIPPED_DIAGONALLY_FLAG   uint32 = 0x20000000
)

type Tile struct {
	GlobalTileID        uint32
	InternalTileID      uint32
	X                   int
	Y                   int
	FlippedHorizontally bool
	FlippedVertically   bool
	FlippedDiagonally   bool
	Tileset             *Tileset
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
	Rendered *ebiten.Image
}

func (l *Layer) DecodeData(gameMap *TmxMap) error {
	if l.Data.Encoding == Base64 {
		byteArray, error := base64.StdEncoding.DecodeString(strings.TrimSpace(l.Data.Text))
		if error != nil {
			return fmt.Errorf("Error decoding data: %s", error)
		}

		tileNum := 0
		for i := 0; i <= len(byteArray)-4; i += 4 {
			newTile := TileFromByteArray(byteArray[i : i+4])

			if newTile.GlobalTileID != 0 {
				for i := range gameMap.Tilesets {
					if newTile.GlobalTileID >= gameMap.Tilesets[i].FirstGid {
						newTile.Tileset = gameMap.Tilesets[i]
					}
				}
				if newTile.Tileset == nil {
					return fmt.Errorf("Couldn't find tileset for %s\n", newTile)
				}

				newTile.X = tileNum % l.Width
				newTile.Y = tileNum / l.Height

				newTile.InternalTileID = newTile.GlobalTileID - newTile.Tileset.FirstGid
				l.Tiles = append(l.Tiles, newTile)
			}

			tileNum++
		}
	}
	return nil
}

func (l *Layer) Render(gameMap *TmxMap, scale float64, refresh bool) *ebiten.Image {
	if l.Rendered == nil || refresh {
		renderStart := time.Now()
		rendered, _ := ebiten.NewImage(gameMap.PixelWidth, gameMap.PixelHeight, ebiten.FilterDefault)
		for _, tile := range l.Tiles {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(tile.X*gameMap.TileWidth), float64(tile.Y*gameMap.TileHeight))
			rendered.DrawImage(tile.Tileset.Tiles[int(tile.InternalTileID)], op)
		}
		l.Rendered = rendered
		t := time.Now()
		elapsed := t.Sub(renderStart)
		fmt.Printf("%s: refreshing layer took %f\n", l.Name, elapsed.Seconds())
	}

	scaledWidth := int(float64(gameMap.CameraBounds.Max.X) / scale)
	scaledHeight := int(float64(gameMap.CameraBounds.Max.Y) / scale)

	gameMap.ScaledCam.Min.X = gameMap.CameraPosition.X - scaledWidth/2
	gameMap.ScaledCam.Min.Y = gameMap.CameraPosition.Y - scaledHeight/2
	gameMap.ScaledCam.Max.X = gameMap.ScaledCam.Min.X + scaledWidth
	gameMap.ScaledCam.Max.Y = gameMap.ScaledCam.Min.Y + scaledHeight

	return l.Rendered.SubImage(gameMap.ScaledCam).(*ebiten.Image)
}

type Object struct {
	Text     string  `xml:",chardata"`
	ID       int     `xml:"id,attr"`
	Name     string  `xml:"name,attr"`
	Type     string  `xml:"type,attr"`
	X        int     `xml:"x,attr"`
	Y        int     `xml:"y,attr"`
	Width    int     `xml:"width,attr"`
	Height   int     `xml:"height,attr"`
	Rotation float64 `xml:"rotation,attr"`
	Gid      uint32  `xml:"gid,attr"`
	Visible  bool    `xml:"visible,attr"`
	Template string  `xml:"template,attr"`
}

type DrawOrder string

const (
	Index   DrawOrder = "index"
	TopDown           = "topdown"
)

type ObjectGroup struct {
	Text      string    `xml:",chardata"`
	ID        int       `xml:"id,attr"`
	Name      string    `xml:"name,attr"`
	Color     string    `xml:"color,attr"`
	X         int       `xml:"x,attr"`
	Y         int       `xml:"y,attr"`
	Width     int       `xml:"width,attr"`
	Height    int       `xml:"height,attr"`
	Opacity   float64   `xml:"opacity,attr"`
	Visible   bool      `xml:"visible,attr"`
	Tintcolor string    `xml:"tintcolor,attr"`
	Offsetx   bool      `xml:"offsetx,attr"`
	OffsetY   bool      `xml:"offsety,attr"`
	DrawOrder DrawOrder `xml:"draworder,attr"`
	Objects   []*Object `xml:"object"`
	Rendered  *ebiten.Image
}

func (o *ObjectGroup) DebugRender(gameMap *TmxMap, scale float64) *ebiten.Image {
	if o.Rendered == nil {
		renderStart := time.Now()
		rendered, _ := ebiten.NewImage(gameMap.PixelWidth, gameMap.PixelHeight, ebiten.FilterDefault)
		for _, obj := range o.Objects {
			objImg, _ := ebiten.NewImage(obj.Width, obj.Height, ebiten.FilterDefault)
			objImg.Fill(image.Black)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(obj.X), float64(obj.Y))
			rendered.DrawImage(objImg, op)
			fmt.Printf("Object: %s, [%d,%d],[%d,%d]\n", obj.Name, obj.X, obj.Y, obj.Width, obj.Height)
		}
		o.Rendered = rendered
		t := time.Now()
		elapsed := t.Sub(renderStart)
		fmt.Printf("%s: refreshing layer took %f\n", o.Name, elapsed.Seconds())
	}
	scaledWidth := int(float64(gameMap.CameraBounds.Max.X) / scale)
	scaledHeight := int(float64(gameMap.CameraBounds.Max.Y) / scale)

	gameMap.ScaledCam.Min.X = gameMap.CameraPosition.X - scaledWidth/2
	gameMap.ScaledCam.Min.Y = gameMap.CameraPosition.Y - scaledHeight/2
	gameMap.ScaledCam.Max.X = gameMap.ScaledCam.Min.X + scaledWidth
	gameMap.ScaledCam.Max.Y = gameMap.ScaledCam.Min.Y + scaledHeight

	return o.Rendered.SubImage(gameMap.ScaledCam).(*ebiten.Image)
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
	PixelWidth       int
	PixelHeight      int
	TileWidth        int            `xml:"tilewidth,attr"`
	TileHeight       int            `xml:"tileheight,attr"`
	HexSideLength    int            `xml:"hexsidelength,attr"`
	StaggerAxis      int            `xml:"staggeraxis,attr"`
	BackgroundColor  string         `xml:"backgroundcolor,attr"`
	Infinite         int            `xml:"infinite,attr"`
	NextLayerID      int            `xml:"nextlayerid,attr"`
	NextObjectID     int            `xml:"nextobjectid,attr"`
	Tilesets         []*Tileset     `xml:"tileset"`
	Layers           []*Layer       `xml:"layer"`
	ObjectGroups     []*ObjectGroup `xml:"objectgroup"`
	CameraPosition   image.Point
	CameraBounds     image.Rectangle
	ScaledCam        image.Rectangle
}

func (t TmxMap) GetObjectGroupByName(name string) *ObjectGroup {
	for i := range t.ObjectGroups {
		if t.ObjectGroups[i].Name == name {
			return t.ObjectGroups[i]
		}
	}
	fmt.Printf("No layer with name '%s' found\n", name)
	return nil
}

func (t TmxMap) CheckColisionPoint(subject image.Point) bool {
	collisionLayer := t.GetObjectGroupByName("collisionmap")

	for _, object := range collisionLayer.Objects {
		if subject.X >= object.X && subject.X <= object.X+object.Width &&
			subject.Y >= object.Y && subject.Y <= object.Y+object.Height {

			fmt.Printf("Collision detected with %s [%d,%d][%d,%d]\n", object.Name, object.X, object.Y, object.Width, object.Height)
			fmt.Printf("Subject [%d,%d]\n", subject.X, subject.Y)
			return true
		}
	}
	return false
}

func (t TmxMap) CheckColision(subject image.Rectangle) bool {
	collisionLayer := t.GetObjectGroupByName("collisionmap")

	for _, object := range collisionLayer.Objects {
		if subject.Min.X < object.X+object.Width &&
			subject.Min.X+subject.Max.X > object.X &&
			subject.Min.Y < object.Y+object.Height &&
			subject.Min.Y+subject.Max.Y > object.Y {

			fmt.Printf("Collision detected with %s [%d,%d][%d,%d]\n", object.Name, object.X, object.Y, object.Width, object.Height)
			fmt.Printf("%s\n", subject)
			return true
		}
	}
	return false
}

func LoadFromFile(path string) (*TmxMap, error) {
	gameMap := &TmxMap{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return gameMap, err
	}

	err = xml.Unmarshal([]byte(data), &gameMap)
	if err != nil {
		return nil, fmt.Errorf("Failed decoding map data: %s", err)
	}

	for i := range gameMap.Tilesets {
		err := gameMap.Tilesets[i].LoadFromTsx(filepath.Dir(path))
		if err != nil {
			return nil, err
		}
	}

	for i := range gameMap.Layers {
		err := gameMap.Layers[i].DecodeData(gameMap)
		if err != nil {
			return nil, fmt.Errorf("Error decoding map layer %d, '%s': %s", i, gameMap.Layers[i].Name, err)
		}
	}

	for _, og := range gameMap.ObjectGroups {
		fmt.Printf("Objectgroup: '%s' with %d objects\n", og.Name, len(og.Objects))
	}
	for i, object := range gameMap.ObjectGroups[0].Objects {
		fmt.Printf("Object #%d: %s [%d/%d, %d/%d]\n", i, object.Name, object.X, object.Y, object.Width, object.Height)
	}

	gameMap.PixelWidth = gameMap.Width * gameMap.TileWidth
	gameMap.PixelHeight = gameMap.Height * gameMap.TileHeight

	return gameMap, nil
}
