package voxel

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/go-gl/mathgl/mgl32"
	lv "github.com/memmaker/voxellib/voxel"
	"github.com/memmaker/voxellib/voxel/anvil"
)

// Map is the battlefield: a box of blocks from (0,0,0) to Size (exclusive)
// in a VoxelLib world, plus the units standing in it.
type Map struct {
	W    *lv.World
	size Int3
	// set changes a block. The client sets it to the engine's SetBlock so
	// the change is relit and remeshed; nil writes the world directly.
	set func(x, y, z int, id lv.StateID) bool

	occupants          map[Int3]MapObject
	knownUnitPositions map[uint64][]Int3
	mapInfoLogger      func(string)
	gameErrorLogger    func(string)

	// cutawayHeight hides all blocks above this Y level (so players can look into buildings); -1 = off
	lit           bool
	cutawayHeight int32
}

// NewMap makes an empty map of the given size in blocks.
func NewMap(reg *lv.Registry, size Int3) *Map {
	w := lv.NewWorld(reg, 0, int(size.Y+15)&^15)
	for cx := int32(0); cx <= (size.X-1)>>4; cx++ {
		for cz := int32(0); cz <= (size.Z-1)>>4; cz++ {
			w.AddChunk(w.NewChunk(lv.ChunkPos{X: int(cx), Z: int(cz)}))
		}
	}
	return OnWorld(w, size)
}

// OnWorld makes a map over an existing world.
func OnWorld(w *lv.World, size Int3) *Map {
	return &Map{W: w, size: size, occupants: map[Int3]MapObject{}, knownUnitPositions: map[uint64][]Int3{}, cutawayHeight: -1}
}

// SetBlockFunc routes block changes through f (see Map.set).
func (m *Map) SetBlockFunc(f func(x, y, z int, id lv.StateID) bool) { m.set = f }

const boundsFile = "bounds.json"

// LoadMap reads a map saved by Save: region files plus its size.
func LoadMap(reg *lv.Registry, dir string) (*Map, error) {
	var size Int3
	b, err := os.ReadFile(filepath.Join(dir, boundsFile))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &size); err != nil {
		return nil, fmt.Errorf("%s: %w", boundsFile, err)
	}
	m := NewMap(reg, size)
	st := anvil.OpenStore(dir, false)
	defer st.Close()
	for _, p := range m.W.Chunks() {
		root, err := st.Read(p.X, p.Z)
		if err != nil {
			return nil, err
		}
		if root == nil {
			continue
		}
		ch, err := anvil.DecodeChunk(m.W, root)
		if err != nil {
			return nil, fmt.Errorf("chunk %v: %w", p, err)
		}
		m.W.RemoveChunk(p)
		m.W.AddChunk(ch)
	}
	return m, nil
}

// Save writes the map's blocks as region files and its size to dir.
func (m *Map) Save(dir string) error {
	if err := os.MkdirAll(filepath.Join(dir, "region"), 0o755); err != nil {
		return err
	}
	b, _ := json.Marshal(m.size)
	if err := os.WriteFile(filepath.Join(dir, boundsFile), b, 0o644); err != nil {
		return err
	}
	st := anvil.OpenStore(dir, true)
	defer st.Close()
	cs := map[[2]int]anvil.Compound{}
	for _, p := range m.W.Chunks() {
		if !m.chunkInBounds(p) {
			continue // lighting margin
		}
		cs[[2]int{p.X, p.Z}] = anvil.EncodeChunk(m.W, m.W.Chunk(p))
	}
	if err := st.WriteAll(cs); err != nil {
		return err
	}
	return st.Sync()
}

func (m *Map) SetLogger(mapLogger func(string), gameErrorLogger func(string)) {
	m.mapInfoLogger = mapLogger
	m.gameErrorLogger = gameErrorLogger
}

func (m *Map) logGameError(msg string) {
	if m.gameErrorLogger != nil {
		m.gameErrorLogger(msg)
	}
}

// GetSize is the map's size in blocks.
func (m *Map) GetSize() Int3 { return m.size }

func (m *Map) Contains(x, y, z int32) bool {
	return x >= 0 && y >= 0 && z >= 0 && x < m.size.X && y < m.size.Y && z < m.size.Z
}

func (m *Map) ContainsGrid(p Int3) bool { return m.Contains(p.X, p.Y, p.Z) }

func (m *Map) ContainsVec(pos mgl32.Vec3) bool { return m.ContainsGrid(PositionToGridInt3(pos)) }

// State is the block state at x,y,z (air outside the map).
func (m *Map) State(x, y, z int32) lv.StateID {
	if !m.Contains(x, y, z) {
		return 0
	}
	return m.W.Block(int(x), int(y), int(z))
}

// Info describes the block at x,y,z.
func (m *Map) Info(x, y, z int32) *lv.StateInfo { return m.W.Reg.Info(m.State(x, y, z)) }

// BlockName is the block's name without properties, e.g. "oak_stairs".
func (m *Map) BlockName(p Int3) string { return m.Info(p.X, p.Y, p.Z).Name }

// IsAir reports an empty or invisible block (air, light).
func (m *Map) IsAir(p Int3) bool { return m.Info(p.X, p.Y, p.Z).Shape == lv.ShapeNone }

// IsSolidBlockAt reports whether the block stops units (outside the map counts as open).
func (m *Map) IsSolidBlockAt(x, y, z int32) bool {
	return m.Contains(x, y, z) && m.Info(x, y, z).Collides
}

func (m *Map) SetBlock(x, y, z int32, id lv.StateID) {
	if !m.Contains(x, y, z) {
		return
	}
	if m.set != nil {
		m.set(int(x), int(y), int(z), id)
	} else {
		m.W.SetBlock(int(x), int(y), int(z), id)
	}
}

func (m *Map) SetAir(p Int3) { m.SetBlock(p.X, p.Y, p.Z, 0) }

func (m *Map) SetFloorAtHeight(y int32, id lv.StateID) {
	for x := int32(0); x < m.size.X; x++ {
		for z := int32(0); z < m.size.Z; z++ {
			m.SetBlock(x, y, z, id)
		}
	}
}

// ChangeCutawayHeight moves the cutaway level by delta blocks. Going above the map turns it off.
// It returns the new level (-1 = off).
func (m *Map) ChangeCutawayHeight(delta int32) int32 {
	top := m.highestSolidLevel()
	current := m.cutawayHeight
	if current < 0 {
		current = top + 1
	}
	current += delta
	if current > top {
		m.cutawayHeight = -1
	} else {
		m.cutawayHeight = max(current, 1)
	}
	return m.cutawayHeight
}

// CutawayHeight returns the highest visible block level, or -1 when everything is visible.
func (m *Map) CutawayHeight() int32 { return m.cutawayHeight }

// CurrentlyDraws reports whether the cutaway shows the block.
func (m *Map) CurrentlyDraws(x, y, z int32) bool {
	return m.cutawayHeight < 0 || y <= m.cutawayHeight
}

func (m *Map) highestSolidLevel() int32 {
	for y := m.size.Y - 1; y > 0; y-- {
		for x := int32(0); x < m.size.X; x++ {
			for z := int32(0); z < m.size.Z; z++ {
				if m.State(x, y, z) != 0 {
					return y
				}
			}
		}
	}
	return 0
}

type MapObject interface {
	GetOccupiedBlockOffsets(Int3) []Int3
	UnitID() uint64
	GetName() string
}

// PositionToGridInt3 converts a Vec3 to a Int3 by flooring the values.
func PositionToGridInt3(pos mgl32.Vec3) Int3 {
	return Int3{int32(math.Floor(float64(pos.X()))), int32(math.Floor(float64(pos.Y()))), int32(math.Floor(float64(pos.Z())))}
}

func DirectionToGridInt3(dir mgl32.Vec3) Int3 {
	return Int3{int32(math.Round(float64(dir.X()))), int32(math.Round(float64(dir.Y()))), int32(math.Round(float64(dir.Z())))}
}

func (m *Map) IsUnitPlaceable(unit MapObject, blockPos Int3) (bool, string) {
	for _, offset := range unit.GetOccupiedBlockOffsets(blockPos) {
		p := blockPos.Add(offset)
		if !m.ContainsGrid(p) {
			return false, "Outside of world"
		}
		if m.IsSolidBlockAt(p.X, p.Y, p.Z) {
			return false, "Wall"
		}
		if m.IsOccupiedExcept(p, unit.UnitID()) {
			b := m.occupants[p]
			return false, fmt.Sprintf("Unit %s(%d) is blocking", b.GetName(), b.UnitID())
		}
	}
	return true, ""
}

func (m *Map) IsHumanoidPlaceable(blockPos Int3) (bool, string) {
	for _, offset := range []Int3{{0, 0, 0}, {0, 1, 0}} {
		p := blockPos.Add(offset)
		if !m.ContainsGrid(p) {
			return false, "Outside of world"
		}
		if m.IsSolidBlockAt(p.X, p.Y, p.Z) {
			return false, "Wall"
		}
	}
	return m.IsSolidBlockAt(blockPos.X, blockPos.Y-1, blockPos.Z), "No floor"
}

func (m *Map) RemoveUnit(unit MapObject) {
	for _, p := range m.knownUnitPositions[unit.UnitID()] {
		if o := m.occupants[p]; o != nil && o.UnitID() == unit.UnitID() {
			delete(m.occupants, p)
		}
	}
	delete(m.knownUnitPositions, unit.UnitID())
}

func (m *Map) SetUnit(unit MapObject, blockPos Int3) bool {
	return m.SetUnitWithOffsets(unit, blockPos, unit.GetOccupiedBlockOffsets(blockPos))
}

func (m *Map) SetUnitWithOffsets(unit MapObject, blockPos Int3, offsets []Int3) bool {
	m.RemoveUnit(unit) // problem: there are cases where we don't place it back
	ok, reason := m.IsUnitPlaceable(unit, blockPos)
	if !ok {
		m.logGameError(fmt.Sprintf("[Map] ERR - Failed to place %s(%d): %s (%s)", unit.GetName(), unit.UnitID(), blockPos.ToString(), reason))
		return false
	}
	occupied := make([]Int3, len(offsets))
	for i, offset := range offsets {
		occupied[i] = blockPos.Add(offset)
		m.occupants[occupied[i]] = unit
	}
	m.knownUnitPositions[unit.UnitID()] = occupied
	return true
}

func (m *Map) GetBlockedCardinalNeighborsUpToHeight(unitID uint64, position Int3, height int32) []Int3 {
	neighbors := make([]Int3, 0, 4)
	for _, offset := range cardinalNeighbors {
		neighbor := position.Add(offset)
		for y := int32(0); y < height; y++ {
			if m.IsSolidBlockAt(neighbor.X, neighbor.Y+y, neighbor.Z) || m.IsOccupiedExcept(neighbor.Add(Int3{0, y, 0}), unitID) {
				neighbors = append(neighbors, neighbor)
				break
			}
		}
	}
	return neighbors
}

func (m *Map) GetNeighborsForGroundMovement(block Int3, keepPredicate func(neighbor Int3) bool) []Int3 {
	neighbors := make([]Int3, 0, 8)
	for _, offset := range cardinalNeighbors {
		neighbor := block.Add(offset)
		if m.ContainsGrid(neighbor) {
			if m.IsSolidBlockAt(neighbor.X, neighbor.Y, neighbor.Z) {
				neighbor = neighbor.Add(Int3{Y: 1}) // climbing one block is allowed
			} else {
				neighbor = m.GetGroundPosition(neighbor)
			}
			if m.ContainsGrid(neighbor) && keepPredicate(neighbor) {
				neighbors = append(neighbors, neighbor)
			}
		}
	}
	for _, offset := range diagonalNeighbors {
		neighbor := m.GetGroundPosition(block.Add(offset))
		if m.ContainsGrid(neighbor) && keepPredicate(neighbor) {
			neighbors = append(neighbors, neighbor)
		}
	}
	return neighbors
}

func (m *Map) GetGroundPosition(startBlock Int3) Int3 {
	for y := startBlock.Y; y >= 1; y-- {
		if m.IsSolidBlockAt(startBlock.X, y-1, startBlock.Z) || !m.ContainsGrid(Int3{startBlock.X, y - 1, startBlock.Z}) {
			return Int3{startBlock.X, y, startBlock.Z}
		}
	}
	return startBlock
}

func (m *Map) IsOccupied(p Int3) bool { return m.occupants[p] != nil }

func (m *Map) IsOccupiedExcept(p Int3, unitID uint64) bool {
	o := m.occupants[p]
	return o != nil && o.UnitID() != unitID
}

func (m *Map) GetMapObjectAt(p Int3) MapObject { return m.occupants[p] }

func (m *Map) IsUnitOnMap(u MapObject) bool {
	_, ok := m.knownUnitPositions[u.UnitID()]
	return ok
}

func (m *Map) DebugGetAllOccupiedBlocks() []Int3 {
	var result []Int3
	for _, blocks := range m.knownUnitPositions {
		result = append(result, blocks...)
	}
	return result
}

func (m *Map) DebugGetOccupiedBlocks(id uint64) []Int3 { return m.knownUnitPositions[id] }

func (m *Map) ForBlockInHalfSphere(origin Int3, radius float64, applyToBlock func(origin Int3, radius float64, x int32, y int32, z int32)) {
	m.forBlockInBall(origin, radius, 0, applyToBlock)
}

func (m *Map) ForBlockInSphere(origin Int3, radius float64, applyToBlock func(origin Int3, radius float64, x int32, y int32, z int32)) {
	m.forBlockInBall(origin, radius, int32(-radius), applyToBlock)
}

func (m *Map) forBlockInBall(origin Int3, radius float64, minY int32, applyToBlock func(origin Int3, radius float64, x int32, y int32, z int32)) {
	for x := int32(-radius); x <= int32(radius); x++ {
		for y := minY; y <= int32(radius); y++ {
			for z := int32(-radius); z <= int32(radius); z++ {
				if float64(x*x+y*y+z*z) <= radius*radius && m.Contains(origin.X+x, origin.Y+y, origin.Z+z) {
					applyToBlock(origin, radius, origin.X+x, origin.Y+y, origin.Z+z)
				}
			}
		}
	}
}

type DistLocPair struct {
	Steps    int
	Distance float64
	Location Int3
}

func (m *Map) FloodFill(origin Int3, maxSteps int, spreadFunc func(origin, location, neighbor Int3, stepsToLocation int) (bool, int)) {
	visited := map[Int3]bool{origin: true}
	queue := []DistLocPair{{0, 0, origin}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.Steps > maxSteps {
			continue
		}
		for _, offset := range cardinalNeighbors3D {
			neighbor := current.Location.Add(offset)
			if !m.ContainsGrid(neighbor) || visited[neighbor] {
				continue
			}
			shouldSpread, neighborSteps := spreadFunc(origin, current.Location, neighbor, current.Steps)
			if !shouldSpread {
				continue
			}
			visited[neighbor] = true
			queue = append(queue, DistLocPair{neighborSteps, origin.Sub(neighbor).Length(), neighbor})
		}
	}
}

func (m *Map) ForBlockInSphericFloodFill(origin Int3, radius float64, applyToBlock func(origin Int3, steps int, distance float64, x int32, y int32, z int32)) {
	visited := map[Int3]bool{origin: true}
	queue := []DistLocPair{{0, 0, origin}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		applyToBlock(origin, current.Steps, current.Distance, current.Location.X, current.Location.Y, current.Location.Z)
		for _, offset := range cardinalNeighbors3D {
			neighbor := current.Location.Add(offset)
			if !m.ContainsGrid(neighbor) || visited[neighbor] || m.IsSolidBlockAt(neighbor.X, neighbor.Y, neighbor.Z) {
				continue
			}
			distance, steps := origin.Sub(neighbor).Length(), current.Steps+1
			if distance > radius || float64(steps) > 2*radius {
				continue
			}
			visited[neighbor] = true
			queue = append(queue, DistLocPair{steps, distance, neighbor})
		}
	}
}

func (m *Map) chunkInBounds(p lv.ChunkPos) bool {
	return p.X >= 0 && p.Z >= 0 && p.X <= int(m.size.X-1)>>4 && p.Z <= int(m.size.Z-1)>>4
}

// Light returns the sky and block light at p, computing the map's light on
// first use. VoxelLib lights a chunk once its eight neighbours exist, so an
// empty ring of chunks is added around the bounds.
func (m *Map) Light(p Int3) (sky, block uint8) {
	if !m.lit {
		m.lit = true
		maxX, maxZ := int(m.size.X-1)>>4, int(m.size.Z-1)>>4
		for cx := -1; cx <= maxX+1; cx++ {
			for cz := -1; cz <= maxZ+1; cz++ {
				if c := (lv.ChunkPos{X: cx, Z: cz}); m.W.Chunk(c) == nil {
					m.W.AddChunk(m.W.NewChunk(c))
				}
			}
		}
		for cx := 0; cx <= maxX; cx++ {
			for cz := 0; cz <= maxZ; cz++ {
				m.W.TryLightChunk(lv.ChunkPos{X: cx, Z: cz})
			}
		}
	}
	return m.W.Light(int(p.X), int(p.Y), int(p.Z))
}
