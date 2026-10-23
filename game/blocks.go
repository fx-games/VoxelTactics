package game

import (
	"sort"

	"github.com/memmaker/voxeltactics/engine/voxel"
	lv "github.com/memmaker/voxellib/voxel"
)

// BlockDefinition is what the game rules add to a block: how it reacts to hits.
type BlockDefinition struct {
	Name                 string
	OnDamageReceived     func(blockPos voxel.Int3, damage int)
	IsBlockingProjectile func() bool
}

var VoidBlockDefinition = &BlockDefinition{Name: "!void"}

// BlockLibrary resolves block names to states and holds the rules per block name.
type BlockLibrary struct {
	Reg  *lv.Registry
	defs map[string]*BlockDefinition
}

func NewBlockLibrary(reg *lv.Registry) *BlockLibrary {
	return &BlockLibrary{Reg: reg, defs: map[string]*BlockDefinition{}}
}

// State is the default state of the named block ("oak_stairs[facing=east]" sets properties).
func (b *BlockLibrary) State(name string) lv.StateID { return b.Reg.Parse(name) }

// Name is the block name of a state, without namespace or properties.
func (b *BlockLibrary) Name(id lv.StateID) string { return shortName(b.Reg.Info(id).Name) }

func shortName(n string) string {
	for i := range n {
		if n[i] == ':' {
			return n[i+1:]
		}
	}
	return n
}

// GetBlockDefinitionByName returns the rules of a block, creating empty ones.
func (b *BlockLibrary) GetBlockDefinitionByName(name string) *BlockDefinition {
	d := b.defs[name]
	if d == nil {
		d = &BlockDefinition{Name: name}
		b.defs[name] = d
	}
	return d
}

// GetBlockDefinition returns the rules of the block a state belongs to.
func (b *BlockLibrary) GetBlockDefinition(id lv.StateID) *BlockDefinition {
	return b.GetBlockDefinitionByName(b.Name(id))
}

func (b *BlockLibrary) ApplyGameplayRules(a *GameInstance) {
	b.GetBlockDefinitionByName("tnt").OnDamageReceived = func(block voxel.Int3, damage int) {
		a.CreateExplodeEffect(block, 4)
	}
	b.GetBlockDefinitionByName("bricks").OnDamageReceived = func(block voxel.Int3, damage int) {
		if damage > 4 {
			a.DestroyBlock(block)
		}
	}
	b.GetBlockDefinitionByName("target").OnDamageReceived = func(blockPos voxel.Int3, damage int) {
		if a.GetMissionDetails().TryDamageObjective(blockPos, damage) {
			a.DestroyBlock(blockPos)
		}
	}
}

func GetDebugBlockNames() []string {
	listOfBlocks := []string{
		"debug2",
		"selection",
		"bricks",
		"clay",
		"copper_block",
		"weathered_copper",
		"weathered_cut_copper",
		"diamond_block",
		"emerald_block",
		"granite",
		"gravel",
		"iron_block",
		"sandstone",
		"ancient_debris",
		"barrel",
		"bedrock",
		"birch_planks",
		"black_terracotta",
		"yellow_terracotta",
		"white_glazed_terracotta",
		"black_wool",
		"brown_terracotta",
		"pink_terracotta",
		"chiseled_quartz_block",
		"cracked_nether_bricks",
		"crafting_table",
		"deepslate_tiles",
		"dispenser",
		"dried_kelp",
		"fletching_table",
		"exposed_copper",
		"furnace",
		"piston",
		"red_nether_bricks",
		"smithing_table",
		"stripped_oak_log",
		"stripped_spruce_log",
		"observer",
		"target",
		"tnt",
	}
	sort.SliceStable(listOfBlocks, func(i, j int) bool {
		return listOfBlocks[i] < listOfBlocks[j]
	})
	return listOfBlocks
}
