package main

import "testing"

func TestGenerationPointsCoverRequestedRadius(t *testing.T) {
	points := generationPoints(16)
	if len(points) != 25 { // -16, -8, 0, 8, 16 in both directions
		t.Fatalf("got %d points, want 25", len(points))
	}
	if points[0] != [2]int{-256, -256} || points[len(points)-1] != [2]int{256, 256} {
		t.Fatalf("unexpected bounds: first=%v last=%v", points[0], points[len(points)-1])
	}
}

func TestGenerationPointsHasMinimumRadius(t *testing.T) {
	if points := generationPoints(0); len(points) == 0 {
		t.Fatal("minimum radius must still produce work")
	}
}
