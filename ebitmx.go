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
}

func (l *Layer) DecodeData(gameMap *TmxMap) error {
	if l.Data.Encoding == Base64 {
		byteArray, error := base64.StdEncoding.DecodeString(strings.TrimSpace(l.Data.Text))
		if error != nil {
			return fmt.Errorf("Error decoding data: %s", error)
		}

		tileNum := 0
		for i := 0; i < len(byteArray)-4; i += 4 {
			newTile := TileFromByteArray(byteArray[i : i+4])

			if newTile.GlobalTileID != 0 {
				for i := range gameMap.Tilesets {
					if newTile.GlobalTileID >= gameMap.Tilesets[i].FirstGid {
						newTile.Tileset = &gameMap.Tilesets[i]
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

func (l Layer) Render(gameMap *TmxMap, camera image.Rectangle, scale float64) *ebiten.Image {
	start := time.Now()

	camWidth := camera.Max.X - camera.Min.X
	camHeight := camera.Max.Y - camera.Min.Y
	widthDiff := int(float64(camWidth)/scale) - camWidth
	heightDiff := int(float64(camHeight)/scale) - camHeight

	//camera.Min.X -= widthDiff / 2
	camera.Max.X += widthDiff
	//camera.Min.Y -= heightDiff / 2
	camera.Max.Y += heightDiff

	rendered, _ := ebiten.NewImage(camWidth+widthDiff, camHeight+heightDiff, ebiten.FilterDefault)

	visibleTiles := getTileRectangleFromAbsolutePixel(camera, l, gameMap.TileWidth, gameMap.TileHeight)
	xOffset := visibleTiles.Min.X*gameMap.TileWidth - camera.Min.X
	yOffset := visibleTiles.Min.Y*gameMap.TileHeight - camera.Min.Y

	startDrawing := time.Now()
	for _, tile := range l.Tiles {
		if tile.X >= visibleTiles.Min.X && tile.X <= visibleTiles.Max.X &&
			tile.Y >= visibleTiles.Min.Y && tile.Y <= visibleTiles.Max.Y {

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(xOffset+tile.X*gameMap.TileWidth), float64(yOffset+tile.Y*gameMap.TileHeight))
			rendered.DrawImage(tile.Tileset.Tiles[int(tile.InternalTileID)], op)
		}
	}
	t := time.Now()
	elapsed := t.Sub(startDrawing)
	fmt.Printf("%s: drawing tiles took %f\n", l.Name, elapsed.Seconds())

	if scale != 1 {
		startScaling := time.Now()

		scaledWidth := int(float64(rendered.Bounds().Max.X) * scale)
		scaledHeight := int(float64(rendered.Bounds().Max.Y) * scale)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)

		scaled, _ := ebiten.NewImage(scaledWidth, scaledHeight, ebiten.FilterDefault)
		scaled.DrawImage(rendered, op)
		rendered = scaled

		t = time.Now()
		elapsed = t.Sub(startScaling)
		fmt.Printf("%s: scaling layer took %f\n", l.Name, elapsed.Seconds())
	}

	t = time.Now()
	elapsed = t.Sub(start)
	fmt.Printf("%s: rendering layer took %f\n", l.Name, elapsed.Seconds())

	return rendered
}
func (l Layer) RenderWHoleLayer(gameMap *TmxMap) *ebiten.Image {
	start := time.Now()

	rendered, _ := ebiten.NewImage(gameMap.PixelWidth, gameMap.PixelHeight, ebiten.FilterDefault)

	for _, tile := range l.Tiles {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(tile.X*gameMap.TileWidth), float64(tile.Y*gameMap.TileHeight))
		rendered.DrawImage(tile.Tileset.Tiles[int(tile.InternalTileID)], op)
	}

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("%s: rendering layer took %f\n", l.Name, elapsed.Seconds())

	return rendered
}

func getTileAbsolutePixelRectangle(tilePosition image.Point, layer Layer) image.Rectangle {
	x0 := tilePosition.X * layer.Width
	x1 := tilePosition.X * (layer.Width + 1)
	y0 := tilePosition.Y * layer.Height
	y1 := tilePosition.Y * (layer.Height + 1)
	return image.Rect(x0, y0, x1, y1)
}

func getTileRectangleFromAbsolutePixel(rectangle image.Rectangle, layer Layer, tileWidth int, tileHeight int) image.Rectangle {
	x0 := rectangle.Min.X / tileWidth
	y0 := rectangle.Min.Y / tileHeight
	x1 := rectangle.Max.X / tileWidth
	y1 := rectangle.Max.Y / tileHeight

	if x0 < 0 {
		x0 = 0
	}
	if x0 > layer.Width {
		x0 = layer.Width
	}
	if y0 < 0 {
		y0 = 0
	}
	if y0 > layer.Height {
		y0 = layer.Height
	}
	if x1 < 0 {
		x1 = 0
	}
	if x1 < layer.Width {
		x1 = layer.Width
	}
	if y1 < 0 {
		y1 = 0
	}
	if y1 < layer.Height {
		y1 = layer.Height
	}
	return image.Rect(x0, y0, x1, y1)
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
	TileWidth        int       `xml:"tilewidth,attr"`
	TileHeight       int       `xml:"tileheight,attr"`
	HexSideLength    int       `xml:"hexsidelength,attr"`
	StaggerAxis      int       `xml:"staggeraxis,attr"`
	BackgroundColor  string    `xml:"backgroundcolor,attr"`
	Infinite         int       `xml:"infinite,attr"`
	NextLayerID      int       `xml:"nextlayerid,attr"`
	NextObjectID     int       `xml:"nextobjectid,attr"`
	Tilesets         []Tileset `xml:"tileset"`
	Layers           []Layer   `xml:"layer"`
}

func LoadFromFile(path string) (*TmxMap, error) {
	gameMap := &TmxMap{}

	data, error := ioutil.ReadFile(path)
	if error != nil {
		return gameMap, error
	}

	_ = xml.Unmarshal([]byte(data), &gameMap)

	for i := range gameMap.Tilesets {
		error = gameMap.Tilesets[i].LoadFromTsx(filepath.Dir(path))
		if error != nil {
			return nil, error
		}
	}

	for i := range gameMap.Layers {
		gameMap.Layers[i].DecodeData(gameMap)
	}

	gameMap.PixelWidth = gameMap.Width * gameMap.TileWidth
	gameMap.PixelWidth = gameMap.Height * gameMap.TileHeight

	return gameMap, error
}
