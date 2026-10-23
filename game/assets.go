package game

import (
	"path"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/memmaker/voxeltactics/engine/afs"
	"github.com/memmaker/voxeltactics/engine/glhf"
	"github.com/memmaker/voxeltactics/engine/util"
	"github.com/memmaker/voxeltactics/engine/voxel"
)

type Assets struct {
	paths map[AssetType]string
}

type AssetType int

const (
	AssetTypeBlockTextures AssetType = iota
	AssetTypeBitmapFonts
	AssetTypeMeshes
	AssetTypeMaps
	AssetTypeSkins
)

func NewAssets() *Assets {
	return &Assets{
		paths: map[AssetType]string{
			AssetTypeBlockTextures: "./assets/textures/blocks/",
			AssetTypeBitmapFonts:   "./assets/fonts/",
			AssetTypeMeshes:        "./assets/models/",
			AssetTypeMaps:          "./assets/maps/",
			AssetTypeSkins:         "./assets/textures/skins/",
		},
	}
}
func (a *Assets) LoadMapWithDetails(mapFile string, details *MissionDetails) *DefaultMapInfo {
	if IsGeneratedMap(mapFile) {
		return a.generateMapWithDetails(mapFile, details)
	}
	mapMetadata := a.LoadMapMetadata(mapFile)
	details.SyncFromMap(mapMetadata)
	library := a.LoadBlockLibrary(mapMetadata.Blocks)
	return &DefaultMapInfo{
		mapFile:      mapFile,
		details:      details,
		metaData:     &mapMetadata,
		blockLibrary: library,
		loadedMap:    a.LoadMap(mapFile),
	}
}

func (a *Assets) generateMapWithDetails(mapName string, details *MissionDetails) *DefaultMapInfo {
	style, seed, err := ParseGeneratedMap(mapName)
	if err != nil {
		panic(err)
	}
	library := a.LoadBlockLibrary(generatedBlockSet)
	generated, mapMetadata, err := GenerateBattlefield(style, seed, library)
	if err != nil {
		panic(err)
	}
	details.SyncFromMap(mapMetadata)
	return &DefaultMapInfo{
		mapFile:      mapName,
		details:      details,
		metaData:     &mapMetadata,
		blockLibrary: library,
		loadedMap:    generated,
	}
}

// LoadVoxelMap returns a fresh copy of the named map, generated or loaded from disk.
func (a *Assets) LoadVoxelMap(mapName string) *voxel.Map {
	if IsGeneratedMap(mapName) {
		return a.generateMapWithDetails(mapName, NewRandomDeathmatch()).GetMap()
	}
	return a.LoadMap(mapName)
}

// LoadBlockLibrary returns a fresh library (rules bind to one game) over the shared registry.
// The name of the map's old texture set is ignored.
func (a *Assets) LoadBlockLibrary(string) *BlockLibrary { return NewBlockLibrary(SharedRegistry()) }

func (a *Assets) LoadBitmapFont(fontName string, glyphWidth int, glyphHeight int) (*glhf.Texture, util.BitmapFontIndex) {
	filePath := path.Join(a.paths[AssetTypeBitmapFonts], fontName)
	fontTextureAtlas := mustLoadTexture(filePath + ".png")
	fontTextureAtlas.SetAtlasItemSize(glyphWidth, glyphHeight)
	atlasIndex := util.NewBitmapFontIndexFromFile(filePath + ".idx")
	return fontTextureAtlas, atlasIndex
}

func (a *Assets) LoadBitmapFontWithoutIndex(fontName string, glyphWidth int, glyphHeight int) *glhf.Texture {
	filePath := path.Join(a.paths[AssetTypeBitmapFonts], fontName)
	fontTextureAtlas := mustLoadTexture(filePath + ".png")
	fontTextureAtlas.SetAtlasItemSize(glyphWidth, glyphHeight)
	return fontTextureAtlas
}

func (a *Assets) LoadMesh(name string) *util.CompoundMesh {
	filePath := a.getModelFile(name)
	return util.LoadGLTFWithTextures(filePath)
}

func (a *Assets) LoadAnimatedMeshWithTextures(name string, animationMap map[string]string) *util.CompoundMesh {
	filePath := a.getModelFile(name)
	return util.LoadGLTFWithAnimationAndTextures(filePath, animationMap)
}
func (a *Assets) LoadMeshWithAnimationMap(name string, animationMap map[string]string) *util.CompoundMesh {
	filePath := a.getModelFile(name)
	return util.LoadGLTF(filePath, animationMap, nil)
}
func (a *Assets) LoadMeshWithoutTextures(name string) *util.CompoundMesh {
	filePath := a.getModelFile(name)
	return util.LoadGLTF(filePath, nil, nil)
}
func (a *Assets) LoadMeshWithColor(name string, forcedColor mgl32.Vec3) *util.CompoundMesh {
	filePath := a.getModelFile(name)
	return util.LoadGLTF(filePath, nil, &forcedColor)
}

func (a *Assets) getModelFile(name string) string {
	binPath := path.Join(a.paths[AssetTypeMeshes], name+".glb")
	if util.DoesFileExist(binPath) {
		return binPath
	}
	return path.Join(a.paths[AssetTypeMeshes], name+".gltf")
}

// LoadMap reads assets/maps/<name>/ (region files and bounds.json).
func (a *Assets) LoadMap(name string) *voxel.Map {
	m, err := voxel.LoadMap(SharedRegistry(), a.GetMapPath(name))
	if err != nil {
		panic(err)
	}
	return m
}

func (a *Assets) LoadMapMetadata(name string) MapMetadata {
	return NewMapMetadataFromFile(path.Join(a.GetMapPath(name), metaFile))
}
func (a *Assets) LoadSkin(file string) *glhf.Texture {
	filePath := path.Join(a.paths[AssetTypeSkins], file+".png")
	return mustLoadTexture(filePath)
}

func (a *Assets) GetMapPath(mapName string) string {
	return path.Join(a.paths[AssetTypeMaps], mapName)
}

func mustLoadTexture(filePath string) *glhf.Texture {
	file, err := afs.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	texture, err := util.NewTextureFromReader(file, false)
	if err != nil {
		panic(err)
	}
	return texture
}
