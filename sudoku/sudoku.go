package sudoku

// Board represents a 9x9 Sudoku board
type Board struct {
	grid [9][9]int
}

// NewBoard creates a new Sudoku board with the given values
func NewBoard(values [9][9]int) *Board {
	return &Board{grid: values}
}

// IsValid checks if a number can be placed at the given position
func (b *Board) IsValid(row, col, num int) bool {
	// Check row
	for x := 0; x < 9; x++ {
		if b.grid[row][x] == num {
			return false
		}
	}

	// Check column
	for x := 0; x < 9; x++ {
		if b.grid[x][col] == num {
			return false
		}
	}

	// Check 3x3 box
	startRow := row - row%3
	startCol := col - col%3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if b.grid[i+startRow][j+startCol] == num {
				return false
			}
		}
	}

	return true
}

// Solve solves the Sudoku puzzle using backtracking
func (b *Board) Solve() bool {
	row, col := b.findEmpty()
	if row == -1 && col == -1 {
		return true // puzzle is solved
	}

	for num := 1; num <= 9; num++ {
		if b.IsValid(row, col, num) {
			b.grid[row][col] = num

			if b.Solve() {
				return true
			}

			b.grid[row][col] = 0 // backtrack
		}
	}

	return false
}

// findEmpty finds an empty cell (represented by 0)
func (b *Board) findEmpty() (int, int) {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if b.grid[i][j] == 0 {
				return i, j
			}
		}
	}
	return -1, -1
}

// GetGrid returns the current state of the grid
func (b *Board) GetGrid() [9][9]int {
	return b.grid
}

// String returns a string representation of the board
func (b *Board) String() string {
	var result string
	for i := 0; i < 9; i++ {
		if i%3 == 0 && i != 0 {
			result += "- - - - - - - - - - - -\n"
		}
		for j := 0; j < 9; j++ {
			if j%3 == 0 && j != 0 {
				result += "| "
			}
			result += string(rune('0'+b.grid[i][j])) + " "
		}
		result += "\n"
	}
	return result
}
