package game

import (
	"encoding/json"
	"os"
	"path"

	"github.com/memmaker/voxeltactics/engine/afs"
	"github.com/memmaker/voxeltactics/engine/util"
	"github.com/memmaker/voxeltactics/engine/voxel"
)

type MapMetadata struct {
	Name           string
	SpawnPositions [][]voxel.Int3
	// PoIPlacements are markers standing on top of each Defend objective block, so the
	// objective itself is one block below (see ObjectivePositions).
	PoIPlacements []voxel.Int3
	Blocks        string // unused: the texture set of the old renderer
	// Lamps are invisible light sources, e.g. inside buildings
	Lamps []voxel.Int3 `json:",omitempty"`
}

// metaFile is the metadata's name inside a map folder.
const metaFile = "meta.json"

// SaveToDisk writes the metadata into the map folder dir.
func (m *MapMetadata) SaveToDisk(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	file, err := os.Create(path.Join(dir, metaFile))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(m)
}

func NewMapMetadataFromFile(filename string) MapMetadata {
	if util.DoesFileExist(filename) {
		var metadata MapMetadata
		data, err := afs.ReadFile(filename)
		if err == nil {
			if util.FromJson(string(data), &metadata) {
				return metadata
			}
		}
	}
	return MapMetadata{
		Name:           "Unnamed Map",
		SpawnPositions: [][]voxel.Int3{{}},
	}
}

// ObjectivePositions returns the Defend objective blocks, one below each PoI marker.
func (m MapMetadata) ObjectivePositions() []voxel.Int3 {
	objectives := make([]voxel.Int3, len(m.PoIPlacements))
	for i, pos := range m.PoIPlacements {
		objectives[i] = pos.Sub(voxel.Int3{Y: 1})
	}
	return objectives
}
