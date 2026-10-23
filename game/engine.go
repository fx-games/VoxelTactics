package game

import "github.com/memmaker/voxeltactics/engine/voxel"

const PositionalTolerance = 0.08

type Engine interface {
	GetVisibleUnits(instance uint64) []*UnitInstance
	GetVoxelMap() *voxel.Map
}
