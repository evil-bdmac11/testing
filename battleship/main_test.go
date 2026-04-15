package main

import (
	"math/rand"
	"testing"
)

func TestParseCoordinate(t *testing.T) {
	tests := []struct {
		in      string
		want    Coord
		wantErr bool
	}{
		{in: "A1", want: Coord{Row: 0, Col: 0}},
		{in: "j10", want: Coord{Row: 9, Col: 9}},
		{in: "K1", wantErr: true},
		{in: "A11", wantErr: true},
		{in: "", wantErr: true},
	}

	for _, tc := range tests {
		got, err := parseCoordinate(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("parseCoordinate(%q): expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseCoordinate(%q): unexpected error %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("parseCoordinate(%q): got %+v, want %+v", tc.in, got, tc.want)
		}
	}
}

func TestBoardPlacementAndSinking(t *testing.T) {
	b := &Board{}
	spec := ShipSpec{Name: "Destroyer", Size: 2, Symbol: 'D'}
	if !b.PlaceShip(spec, Coord{Row: 0, Col: 0}, true) {
		t.Fatal("expected valid placement")
	}
	if b.PlaceShip(spec, Coord{Row: 0, Col: 1}, false) {
		t.Fatal("expected overlapping placement to fail")
	}

	hit, sunk, already := b.ReceiveShot(Coord{Row: 0, Col: 0})
	if already || !hit || sunk != nil {
		t.Fatalf("first hit wrong state: hit=%v sunk=%v already=%v", hit, sunk != nil, already)
	}
	hit, sunk, already = b.ReceiveShot(Coord{Row: 0, Col: 1})
	if already || !hit || sunk == nil {
		t.Fatalf("second hit should sink ship: hit=%v sunk=%v already=%v", hit, sunk != nil, already)
	}
	if !b.AllSunk() {
		t.Fatal("expected all ships sunk")
	}
}

func TestAITargetsAdjacentAfterHit(t *testing.T) {
	ai := NewAI(rand.New(rand.NewSource(1)))
	hitCell := Coord{Row: 5, Col: 5}
	ai.RegisterResult(hitCell, true, false)
	next := ai.NextShot()
	dr := next.Row - hitCell.Row
	if dr < 0 {
		dr = -dr
	}
	dc := next.Col - hitCell.Col
	if dc < 0 {
		dc = -dc
	}
	if dr+dc != 1 {
		t.Fatalf("expected adjacent follow-up shot, got %+v", next)
	}
}

func TestAINoRepeatedShots(t *testing.T) {
	ai := NewAI(rand.New(rand.NewSource(2)))
	seen := map[Coord]bool{}
	for i := 0; i < boardSize*boardSize; i++ {
		n := ai.NextShot()
		if seen[n] {
			t.Fatalf("AI repeated shot at %+v", n)
		}
		seen[n] = true
		ai.RegisterResult(n, false, false)
	}
}
