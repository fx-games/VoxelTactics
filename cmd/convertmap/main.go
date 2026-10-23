// Command convertmap turns the old map formats (gzip chunk dumps with byte block
// IDs, and .construction files) into region worlds under assets/maps/<name>/.
//
//	go run ./cmd/convertmap assets/maps/map.bin assets/maps/paris.construction
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	lv "github.com/memmaker/voxellib/voxel"
	"github.com/memmaker/voxeltactics/engine/util"
	"github.com/memmaker/voxeltactics/engine/voxel"
	"github.com/memmaker/voxeltactics/game"
)

func main() {
	library := game.NewBlockLibrary(game.SharedRegistry())
	for _, src := range os.Args[1:] {
		var m *voxel.Map
		var err error
		if strings.HasSuffix(src, ".construction") {
			m = convertConstruction(library, src)
		} else {
			m, err = convertBin(library, src)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, src, err)
			os.Exit(1)
		}
		name := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(src), ".bin"), ".construction")
		dir := filepath.Join(filepath.Dir(src), name)
		if err := m.Save(dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		meta := game.MapMetadata{Name: name}
		if _, err := os.Stat(src + ".meta"); err == nil {
			meta = game.NewMapMetadataFromFile(src + ".meta")
		}
		if err := meta.SaveToDisk(dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(src, "->", dir, m.GetSize())
	}
}

// convertBin reads the old chunk dump: block IDs are 1-based positions in the
// star_odyssey_01 block list (0 is air).
func convertBin(library *game.BlockLibrary, src string) (*voxel.Map, error) {
	raw, err := os.ReadFile(src)
	if err != nil {
		return nil, err
	}
	r, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	var hdr struct{ Horizontal, Height, W, H, D int32 }
	var count int16
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	names := util.NewBlockListFromFile("assets/textures/blocks/star_odyssey_01.txt")
	states := make([]lv.StateID, len(names)+1)
	for i, n := range names {
		if n != "debug2" && n != "selection" { // editor markers, not blocks
			states[i+1] = library.State(n)
		}
	}
	h, ch := hdr.Horizontal, hdr.Height
	m := voxel.NewMap(library.Reg, voxel.Int3{X: hdr.W * h, Y: hdr.H * ch, Z: hdr.D * h})
	data := make([]byte, h*ch*h)
	for i := int16(0); i < count; i++ {
		var pos [3]int32
		if err := binary.Read(r, binary.LittleEndian, &pos); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}
		for j, id := range data {
			if id == 0 {
				continue
			}
			if int(id) >= len(states) {
				return nil, fmt.Errorf("block id %d outside the block list", id)
			}
			x, y, z := int32(j)%h, int32(j)/h%ch, int32(j)/(h*ch)
			m.SetBlock(pos[0]*h+x, pos[1]*ch+y, pos[2]*h+z, states[id])
		}
	}
	return m, nil
}

// convertConstruction places a .construction file's blocks, keeping their
// properties, with its minimum corner at the origin.
func convertConstruction(library *game.BlockLibrary, src string) *voxel.Map {
	c := voxel.LoadConstruction(src)
	lo := voxel.Int3{X: 1 << 30, Y: 1 << 30, Z: 1 << 30}
	hi := voxel.Int3{X: -1 << 30, Y: -1 << 30, Z: -1 << 30}
	for _, s := range c.Sections {
		lo = voxel.Int3{X: min(lo.X, s.MinBlockX), Y: min(lo.Y, s.MinBlockY), Z: min(lo.Z, s.MinBlockZ)}
		hi = voxel.Int3{X: max(hi.X, s.MinBlockX+int32(s.ShapeX)), Y: max(hi.Y, s.MinBlockY+int32(s.ShapeY)), Z: max(hi.Z, s.MinBlockZ+int32(s.ShapeZ))}
	}
	m := voxel.NewMap(library.Reg, hi.Sub(lo))
	for _, s := range c.Sections {
		i := 0
		for x := int32(0); x < int32(s.ShapeX); x++ {
			for y := int32(0); y < int32(s.ShapeY); y++ {
				for z := int32(0); z < int32(s.ShapeZ); z++ {
					b := s.Blocks[i]
					i++
					if b == nil || b.Name == "air" {
						continue
					}
					props := map[string]string{}
					for k, v := range b.Properties {
						props[k] = fmt.Sprint(v)
					}
					p := voxel.Int3{X: s.MinBlockX + x, Y: s.MinBlockY + y, Z: s.MinBlockZ + z}.Sub(lo)
					m.SetBlock(p.X, p.Y, p.Z, library.Reg.State(b.Name, props))
				}
			}
		}
	}
	return m
}
