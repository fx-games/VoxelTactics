package voxel

var cardinalNeighbors = []Int3{
	{1, 0, 0},
	{-1, 0, 0},
	{0, 0, 1},
	{0, 0, -1},
}

var cardinalNeighborsUpDown = []Int3{
	{0, 1, 0},
	{0, -1, 0},
}
var cardinalNeighbors3D = append(cardinalNeighbors, cardinalNeighborsUpDown...)
var diagonalNeighbors = []Int3{
	{1, 0, 1},
	{-1, 0, 1},
	{1, 0, -1},
	{-1, 0, -1},
}

var allNeighbors = append(cardinalNeighbors, diagonalNeighbors...)

var NorthDir = Int3{X: 0, Y: 0, Z: -1}
var SouthDir = Int3{X: 0, Y: 0, Z: 1}
var EastDir = Int3{X: 1, Y: 0, Z: 0}
var WestDir = Int3{X: -1, Y: 0, Z: 0}
var Up = Int3{X: 0, Y: 1, Z: 0}
var Down = Int3{X: 0, Y: -1, Z: 0}
