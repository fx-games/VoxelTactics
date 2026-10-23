module github.com/memmaker/voxeltactics

go 1.25.0

require (
	github.com/Tnze/go-mc v1.19.3
	github.com/faiface/mainthread v0.0.0-20171120011319-8b78f0a41ae3
	github.com/fatih/color v1.15.0
	github.com/go-gl/gl v0.0.0-20211210172815-726fda9656d6
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20250301202403-da16c1255728
	github.com/go-gl/mathgl v1.2.0
	github.com/ojrac/opensimplex-go v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/qmuntal/gltf v0.24.2
	github.com/solarlune/gocoro v0.0.0-20230824040254-0c0c4a19d76f
	github.com/spakin/netpbm v1.3.0
)

require (
	github.com/ebitengine/oto/v3 v3.3.3
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/memmaker/voxellib v0.0.0
	golang.org/x/sys v0.47.0 // indirect
)

replace github.com/memmaker/voxellib => ../VoxelLib
