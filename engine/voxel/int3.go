package voxel

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

type Int3 struct {
	X, Y, Z int32
}

func (i Int3) Add(other Int3) Int3 {
	return Int3{X: i.X + other.X, Y: i.Y + other.Y, Z: i.Z + other.Z}
}

func (i Int3) Mul(factor int32) Int3 {
	i.X *= factor
	i.Y *= factor
	i.Z *= factor
	return i
}

func (i Int3) ToFloatVec3() mgl32.Vec3 {
	return mgl32.Vec3{float32(i.X), float32(i.Y), float32(i.Z)}
}

func (i Int3) Sub(tr Int3) Int3 {
	return Int3{X: i.X - tr.X, Y: i.Y - tr.Y, Z: i.Z - tr.Z}
}

func (i Int3) ToVec3() mgl32.Vec3 {
	return mgl32.Vec3{float32(i.X), float32(i.Y), float32(i.Z)}
}

func (i Int3) Div(factor int) Int3 {
	return Int3{X: i.X / int32(factor), Y: i.Y / int32(factor), Z: i.Z / int32(factor)}
}

func (i Int3) ToBlockCenterVec3() mgl32.Vec3 {
	return mgl32.Vec3{float32(i.X) + 0.5, float32(i.Y), float32(i.Z) + 0.5}
}

func (i Int3) ToBlockCenterVec3D() mgl32.Vec3 {
	return mgl32.Vec3{float32(i.X) + 0.5, float32(i.Y), float32(i.Z) + 0.5}
}

func (i Int3) ToString() string {
	return fmt.Sprintf("(%d,%d,%d)", i.X, i.Y, i.Z)
}
func (i Int3) ToCardinalDirection() Int3 {
	absX := int32(math.Abs(float64(i.X)))
	absZ := int32(math.Abs(float64(i.Z)))

	signX := int32(1)
	signZ := int32(1)

	if i.X < 0 {
		signX = -1
	}
	if i.Z < 0 {
		signZ = -1
	}

	if absX > absZ {
		return Int3{X: signX, Y: 0, Z: 0}
	} else {
		return Int3{X: 0, Y: 0, Z: signZ}
	}
}
func (i Int3) ToDiagonalDirection() Int3 {
	// allowed values are
	// north: 0,0,-1
	// east: 1,0,0
	// south: 0,0,1
	// west: -1,0,0
	// north-east: 1,0,-1
	// south-east: 1,0,1
	// south-west: -1,0,1
	// north-west: -1,0,-1
	absX := int32(math.Abs(float64(i.X)))
	absZ := int32(math.Abs(float64(i.Z)))
	signX := int32(1)
	signZ := int32(1)
	if i.X < 0 {
		signX = -1
	}
	if i.Z < 0 {
		signZ = -1
	}

	ratio := float64(absX) / float64(absZ)

	if ratio > 0.5 && ratio < 1.5 { // diagonal
		return Int3{X: signX, Y: 0, Z: signZ}
	} else if ratio >= 1.5 { // vertical
		return Int3{X: signX, Y: 0, Z: 0}
	} else { // horizontal
		return Int3{X: 0, Y: 0, Z: signZ}
	}
}

func (i Int3) ManhattanLength() int {
	return int(math.Abs(float64(i.X)) + math.Abs(float64(i.Y)) + math.Abs(float64(i.Z)))
}

func (i Int3) ManhattanLength2() int {
	return int(math.Abs(float64(i.X)) + math.Abs(float64(i.Z)))
}

func (i Int3) IsBelow(grid Int3) bool {
	return i.Y == grid.Y-1
}

func (i Int3) LengthInt() int32 {
	return int32(math.Sqrt(float64(i.X*i.X + i.Y*i.Y + i.Z*i.Z)))
}
func (i Int3) Length() float64 {
	return math.Sqrt(float64(i.X*i.X + i.Y*i.Y + i.Z*i.Z))
}
