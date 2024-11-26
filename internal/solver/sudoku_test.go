package solver

import (
	"testing"
)

func TestNewBoard(t *testing.T) {
	values := [9][9]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}

	board := NewBoard(values)
	if board == nil {
		t.Error("NewBoard returned nil")
	}

	if board.grid != values {
		t.Error("NewBoard did not correctly initialize the grid")
	}
}

func TestIsValid(t *testing.T) {
	board := &Board{}

	// Test empty board
	if !board.IsValid(0, 0, 1) {
		t.Error("IsValid should return true for valid move on empty board")
	}

	// Test row conflict
	board.grid[0][0] = 1
	if board.IsValid(0, 1, 1) {
		t.Error("IsValid should return false for row conflict")
	}

	// Test column conflict
	board.grid[0][0] = 1
	if board.IsValid(1, 0, 1) {
		t.Error("IsValid should return false for column conflict")
	}

	// Test 3x3 box conflict
	board.grid[0][0] = 1
	if board.IsValid(1, 1, 1) {
		t.Error("IsValid should return false for box conflict")
	}
}

func TestSolve(t *testing.T) {
	tests := []struct {
		name     string
		input    [9][9]int
		solvable bool
	}{
		{
			name: "Valid puzzle",
			input: [9][9]int{
				{5, 3, 0, 0, 7, 0, 0, 0, 0},
				{6, 0, 0, 1, 9, 5, 0, 0, 0},
				{0, 9, 8, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{4, 0, 0, 8, 0, 3, 0, 0, 1},
				{7, 0, 0, 0, 2, 0, 0, 0, 6},
				{0, 6, 0, 0, 0, 0, 2, 8, 0},
				{0, 0, 0, 4, 1, 9, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 7, 9},
			},
			solvable: true,
		},
		{
			name: "Empty puzzle",
			input: [9][9]int{
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
			solvable: true,
		},
		{
			name: "Unsolvable puzzle",
			input: [9][9]int{
				{5, 3, 0, 0, 7, 0, 0, 0, 0},
				{5, 0, 0, 1, 9, 5, 0, 0, 0}, // Note the duplicate 5 in first column
				{0, 9, 8, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{4, 0, 0, 8, 0, 3, 0, 0, 1},
				{7, 0, 0, 0, 2, 0, 0, 0, 6},
				{0, 6, 0, 0, 0, 0, 2, 8, 0},
				{0, 0, 0, 4, 1, 9, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 7, 9},
			},
			solvable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := NewBoard(tt.input)
			result := board.Solve()
			if result != tt.solvable {
				t.Errorf("Solve() = %v, want %v", result, tt.solvable)
			}

			if result {
				// Verify solution is valid
				for i := 0; i < 9; i++ {
					for j := 0; j < 9; j++ {
						if board.grid[i][j] == 0 {
							t.Error("Solution contains empty cells")
						}
					}
				}
			}
		})
	}
}

func TestFindEmpty(t *testing.T) {
	tests := []struct {
		name     string
		board    *Board
		wantRow  int
		wantCol  int
	}{
		{
			name: "No empty cells",
			board: NewBoard([9][9]int{
				{1, 2, 3, 4, 5, 6, 7, 8, 9},
				{4, 5, 6, 7, 8, 9, 1, 2, 3},
				{7, 8, 9, 1, 2, 3, 4, 5, 6},
				{2, 3, 1, 5, 6, 4, 8, 9, 7},
				{5, 6, 4, 8, 9, 7, 2, 3, 1},
				{8, 9, 7, 2, 3, 1, 5, 6, 4},
				{3, 1, 2, 6, 4, 5, 9, 7, 8},
				{6, 4, 5, 9, 7, 8, 3, 1, 2},
				{9, 7, 8, 3, 1, 2, 6, 4, 5},
			}),
			wantRow: -1,
			wantCol: -1,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRow, gotCol := tt.board.findEmpty()
			if gotRow != tt.wantRow || gotCol != tt.wantCol {
				t.Errorf("findEmpty() = (%v, %v), want (%v, %v)", 
					gotRow, gotCol, tt.wantRow, tt.wantCol)
			}
		})
	}
}

func TestString(t *testing.T) {
	board := NewBoard([9][9]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	})

	result := board.String()
	if result == "" {
		t.Error("String() returned empty string")
	}
}
