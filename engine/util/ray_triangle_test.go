package util

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func TestSegmentStopsAtItsEnd(t *testing.T) {
	a, b, c := mgl32.Vec3{0, 0, 5}, mgl32.Vec3{1, 0, 5}, mgl32.Vec3{0, 1, 5}
	if hit, _ := intersectLineSegmentTriangle(mgl32.Vec3{0.2, 0.2, 0}, mgl32.Vec3{0.2, 0.2, 4}, a, b, c); hit {
		t.Error("a segment ending before the triangle must not hit it")
	}
	if hit, p := intersectLineSegmentTriangle(mgl32.Vec3{0.2, 0.2, 0}, mgl32.Vec3{0.2, 0.2, 6}, a, b, c); !hit || p.Z() != 5 {
		t.Errorf("expected a hit at z=5, got %v %v", hit, p)
	}
}
