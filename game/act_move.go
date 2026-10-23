package game

import (
	"slices"

	"github.com/memmaker/voxeltactics/engine/voxel"
)

type ActionMove struct {
	gameMap         *voxel.Map
	selectedPath    []voxel.Int3
	previousNodeMap map[voxel.Int3]voxel.Int3
	distanceMap     map[voxel.Int3]float64
	unit            *UnitInstance
	validTargets    []voxel.Int3
}

func (a *ActionMove) IsValidTarget(target voxel.Int3) bool {
	distance, ok := a.distanceMap[target]
	return ok && distance <= float64(a.unit.MovesLeft())
}

func NewActionMove(gameMap *voxel.Map, unit *UnitInstance) *ActionMove {
	a := &ActionMove{
		gameMap:         gameMap,
		previousNodeMap: make(map[voxel.Int3]voxel.Int3),
		distanceMap:     make(map[voxel.Int3]float64),
		unit:            unit,
	}
	a.updateTargetData()
	return a
}

func (a *ActionMove) GetName() string {
	return "Move"
}

func (a *ActionMove) GetValidTargets() []voxel.Int3 {
	return a.validTargets
}

func (a *ActionMove) GetPath(target voxel.Int3) []voxel.Int3 {
	pathToTarget := make([]voxel.Int3, 0)
	current := target
	index := 0
	for {
		pathToTarget = append(pathToTarget, current)
		if prev, ok := a.previousNodeMap[current]; ok {
			current = prev
			index++
		} else {
			break
		}
	}
	// remove last element, which is the current position
	pathToTarget = pathToTarget[:len(pathToTarget)-1]
	// reverse
	slices.Reverse(pathToTarget)
	return pathToTarget
}

func (a *ActionMove) GetCost(target voxel.Int3) float64 {
	return a.distanceMap[target]
}

func (a *ActionMove) updateTargetData() {
	footPosInt := a.unit.GetBlockPosition()
	var valid []voxel.Int3
	cost, prev := NewMoveGraph(a.gameMap, a.unit).Reach(toPos(footPosInt), int(a.unit.MovesLeft()*navCost))
	for p, c := range cost {
		node := toInt3(p)
		if node == footPosInt {
			continue
		}
		valid = append(valid, node)
		a.distanceMap[node] = float64(c) / navCost
	}
	for p, from := range prev {
		a.previousNodeMap[toInt3(p)] = toInt3(from)
	}
	a.validTargets = valid
}
