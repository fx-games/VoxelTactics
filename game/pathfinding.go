package game

import (
	"github.com/memmaker/voxellib/voxel/nav"
	"github.com/memmaker/voxeltactics/engine/voxel"
)

// navCost is VoxelLib's cost of one straight step; move budgets are in steps.
const navCost = 10

// solidMap is the map as nav sees it: outside the bounds is solid, so walkers stay on the map.
type solidMap struct{ m *voxel.Map }

func (s solidMap) Solid(x, y, z int) bool {
	if !s.m.Contains(int32(x), int32(y), int32(z)) {
		return true
	}
	return s.m.IsSolidBlockAt(int32(x), int32(y), int32(z))
}

// NewMoveGraph is the walking graph for a unit; occupancy and stance footprint via Passable.
func NewMoveGraph(voxelMap *voxel.Map, unit *UnitInstance) *nav.Graph {
	return &nav.Graph{W: solidMap{voxelMap}, Passable: func(_, to nav.Pos) bool {
		ok, _ := voxelMap.IsUnitPlaceable(unit, toInt3(to))
		return ok
	}}
}

func toPos(p voxel.Int3) nav.Pos  { return nav.Pos{X: int(p.X), Y: int(p.Y), Z: int(p.Z)} }
func toInt3(p nav.Pos) voxel.Int3 { return voxel.Int3{X: int32(p.X), Y: int32(p.Y), Z: int32(p.Z)} }
