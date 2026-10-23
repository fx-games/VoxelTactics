package game

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	lv "github.com/memmaker/voxellib/voxel"
	"github.com/memmaker/voxellib/voxel/mcres"
)

// BasePackPath is the pack that gives blocks their shapes: $VT_BASE_PACK, else
// the newest release client of the installed block game. "" when none is found.
func BasePackPath() string {
	if p := os.Getenv("VT_BASE_PACK"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Library", "Application Support", "minecraft", "versions")
	entries, _ := os.ReadDir(dir)
	var releases []string
	for _, e := range entries {
		if n := e.Name(); e.IsDir() && !strings.Contains(n, "-") {
			releases = append(releases, n)
		}
	}
	sort.Slice(releases, func(i, j int) bool { return versionLess(releases[i], releases[j]) })
	for i := len(releases) - 1; i >= 0; i-- {
		if p := filepath.Join(dir, releases[i], releases[i]+".jar"); fileExists(p) {
			return p
		}
	}
	return ""
}

// versionLess compares dotted version numbers ("26.3" < "26.10").
func versionLess(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		if len(as[i]) != len(bs[i]) {
			return len(as[i]) < len(bs[i])
		}
		if as[i] != bs[i] {
			return as[i] < bs[i]
		}
	}
	return len(as) < len(bs)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

var shared struct {
	once sync.Once
	reg  *lv.Registry
}

// SharedRegistry is the process's block registry (registries are safe for concurrent use).
func SharedRegistry() *lv.Registry {
	shared.once.Do(func() { shared.reg, _ = NewRegistry() })
	return shared.reg
}

// NewRegistry makes the block registry the server and client share. With a
// base pack, unknown blocks get its shapes; the returned pack (nil without
// one) must stay open while the registry resolves blocks.
func NewRegistry() (*lv.Registry, *mcres.Pack) {
	reg := lv.NewRegistry()
	p := BasePackPath()
	if p == "" {
		println("[Blocks] No base pack found (set VT_BASE_PACK): unknown blocks are plain cubes")
		return reg, nil
	}
	pack, err := mcres.Open(p)
	if err != nil {
		println("[Blocks] Base pack: " + err.Error())
		return reg, nil
	}
	reg.Resolver = pack.Resolver()
	reg.Faces = pack.FaceResolver()
	return reg, pack
}
