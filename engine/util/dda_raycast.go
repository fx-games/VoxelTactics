package util

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	lv "github.com/memmaker/voxellib/voxel"
	"github.com/memmaker/voxellib/voxel/physics"
	"github.com/memmaker/voxeltactics/engine/voxel"
)

type CardinalDirection int

const (
	North CardinalDirection = iota
	East
	South
	West
)

type HitInfo2D struct {
	Distance               float64
	Side                   CardinalDirection
	CollisionX, CollisionY float64
	GridX                  int64
	GridY                  int64
}

func Raycast2D(startX, startY, directionX, directionY float64, shouldStopRay func(x, y int64) bool) HitInfo2D {
	// 2. Die sich wiederholenden Strahlenteile für Schritte in CollisionX und CollisionY Richtung berechnen
	//    - Steigung des Richtungsvektors und ein Schritt in CollisionX bzw. CollisionY Richtung
	//    - Satz des Pythagoras mit a = 1 und b = ray.y / ray.x für einen Schritt in die CollisionX-Richtung
	//    - Satz des Pythagoras mit a = 1 und b = ray.x / ray.y für einen Schritt in die CollisionY-Richtung
	deltaDistX := math.Sqrt(1 + (directionY*directionY)/(directionX*directionX))
	deltaDistY := math.Sqrt(1 + (directionX*directionX)/(directionY*directionY))

	startGridX := int64(math.Floor(startX))
	startGridY := int64(math.Floor(startY))

	mapX := startGridX
	mapY := startGridY

	var intraCellPositionX float64
	var intraCellPositionY float64

	var sideDistX float64
	var sideDistY float64

	var mapStepX int64
	var mapStepY int64

	// 3. Anhand der Richtung des Strahls folgende Werte berechnen
	//     - Die Schrittrichtung auf der Karte (je Quadrant)
	//     - Die translation innerhalb der aktuellen Zelle
	//     - Die anteiligen Startteile der beiden Strahlen für CollisionX und CollisionY Schritte
	if directionX < 0 {
		mapStepX = -1
		intraCellPositionX = startX - float64(mapX)
		sideDistX = intraCellPositionX * deltaDistX
	} else {
		mapStepX = 1
		intraCellPositionX = float64(mapX) + 1.0 - startX
		sideDistX = intraCellPositionX * deltaDistX
	}

	if directionY < 0 {
		mapStepY = -1
		intraCellPositionY = startY - float64(mapY)
		sideDistY = intraCellPositionY * deltaDistY
	} else {
		mapStepY = 1
		intraCellPositionY = float64(mapY) + 1.0 - startY
		sideDistY = intraCellPositionY * deltaDistY
	}

	eastWestSide := false
	stopRaycasting := false
	nextCollisionX := 0.0
	nextCollisionY := 0.0

	// 4. Schrittweise die Strahlen verlängern, immer der kürzeste zuerst
	//    - Die Kartenposition wird aktualisiert
	//    - Die Wandrichtung (Nord/Süd vs. Ost/west) wird gesetzt
	//    - Der Strahl wird verlängert
	for !stopRaycasting {
		if sideDistX < sideDistY {
			// move one unit in CollisionX Direction
			nextCollisionX = startX + (directionX * sideDistX)
			nextCollisionY = startY + (directionY * sideDistX)

			mapX += mapStepX
			eastWestSide = true
			sideDistX += deltaDistX
		} else {
			// move one unit in CollisionY Direction
			nextCollisionX = startX + (directionX * sideDistY)
			nextCollisionY = startY + (directionY * sideDistY)

			mapY += mapStepY
			eastWestSide = false
			sideDistY += deltaDistY
		}

		// Optional: Sprites - Prüfen ob Sprites auf diesem Feld sind
		//Sprite visibleSprite = GetSpriteMapAt(mapX, mapY);
		/*
		   if (visibleSprite != null && !mVisibleSprites.Contains(visibleSprite))
		       mVisibleSprites.AddFlat(visibleSprite);
		*/

		// 5. Prüfen wir ob eine Wand getroffen wurde
		stopRaycasting = shouldStopRay(mapX, mapY)
	}

	// 6. Den Abstand zur Wand berechnen (Im Prinzip den letzten Schritt rückgängig machen)
	// Vektoren subtrahieren
	var perpWallDist float64
	if eastWestSide {
		perpWallDist = sideDistX - deltaDistX
	} else {
		perpWallDist = sideDistY - deltaDistY
	}

	// 6.1 Die original Distanz zur Wand für das Texture Mapping verwenden
	var wallX float64 // CollisionX translation an der wir die Wand getroffen haben
	// Komponentenweise Vektoraddition und Skalierung
	if eastWestSide {
		wallX = startY + (perpWallDist * directionY)
	} else {
		wallX = startX + (perpWallDist * directionX)
	}

	wallX -= math.Floor(wallX) // Uns interessieren nur die Nachkommastellen
	//textureX := (int)(wallX * mWallTextures[textureIndex].Width);

	var hitSide CardinalDirection
	if eastWestSide {
		if directionX < 0 {
			hitSide = West
		} else {
			hitSide = East
		}
	} else {
		if directionY < 0 {
			hitSide = North
		} else {
			hitSide = South
		}
	}
	return HitInfo2D{
		Distance:   perpWallDist,
		Side:       hitSide,
		CollisionX: nextCollisionX,
		CollisionY: nextCollisionY,
		GridX:      mapX,
		GridY:      mapY,
	}
}

type CubeSide int

const (
	Front CubeSide = iota
	Back
	Left
	Right
	Top
	Bottom
)

type HitInfo3D struct {
	Distance               float64
	Side                   CubeSide
	CollisionWorldPosition mgl32.Vec3
	PreviousGridPosition   voxel.Int3
	CollisionGridPosition  voxel.Int3
	Hit                    bool
}

func (d HitInfo3D) WithCollisionWorldPosition(point mgl32.Vec3) HitInfo3D {
	d.CollisionWorldPosition = point
	return d
}

func (d HitInfo3D) WithDistance(newDist float64) HitInfo3D {
	d.Distance = newDist
	return d
}

// DDARaycast walks the grid cells from rayStart to rayEnd (VoxelLib physics.Walk)
// until stopRay returns true.
func DDARaycast(rayStart, rayEnd mgl32.Vec3, stopRay func(x, y, z int32) bool) HitInfo3D {
	ray := rayEnd.Sub(rayStart)
	length := float64(ray.Len())
	if length == 0 {
		return HitInfo3D{}
	}
	dir := ray.Mul(float32(1 / length))
	var hit HitInfo3D
	physics.Walk(nil, float64(rayStart.X()), float64(rayStart.Y()), float64(rayStart.Z()), float64(dir.X()), float64(dir.Y()), float64(dir.Z()), length,
		func(x, y, z int, _ lv.StateID, f lv.Face, t float64) bool {
			if !stopRay(int32(x), int32(y), int32(z)) {
				return false
			}
			cell := voxel.Int3{X: int32(x), Y: int32(y), Z: int32(z)}
			side, back := faceSide[f].side, faceSide[f].back
			if t == 0 { // the ray starts inside the cell
				side, back = Top, voxel.Int3{}
			}
			hit = HitInfo3D{
				Hit:                    true,
				Distance:               t,
				Side:                   side,
				CollisionWorldPosition: rayStart.Add(dir.Mul(float32(t))),
				PreviousGridPosition:   cell.Add(back),
				CollisionGridPosition:  cell,
			}
			return true
		})
	return hit
}

// faceSide maps the face a ray enters through to the old side names and the
// offset back to the cell it came from.
var faceSide = [lv.FaceCount]struct {
	side CubeSide
	back voxel.Int3
}{
	lv.FaceDown:  {Bottom, voxel.Int3{Y: -1}},
	lv.FaceUp:    {Top, voxel.Int3{Y: 1}},
	lv.FaceNorth: {Back, voxel.Int3{Z: -1}},
	lv.FaceSouth: {Front, voxel.Int3{Z: 1}},
	lv.FaceWest:  {Left, voxel.Int3{X: -1}},
	lv.FaceEast:  {Right, voxel.Int3{X: 1}},
}
