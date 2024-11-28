package solver

import (
	"fmt"
	"math/rand"
)

// Board represents a 9x9 Sudoku board
type Board struct {
	grid [9][9]int
}

// NewBoard creates a new Sudoku board with the given values
func NewBoard(values [9][9]int, validate bool) (*Board, error) {
	board := &Board{grid: values}
	if validate && !board.isValidBoard() {
		return nil, fmt.Errorf("invalid board: contains duplicate numbers or invalid values")
	}
	return board, nil
}

// isValidBoard checks if the current board state is valid
func (b *Board) isValidBoard() bool {
	// Check each row and column
	for i := 0; i < 9; i++ {
		rowNums := make(map[int]bool)
		colNums := make(map[int]bool)

		for j := 0; j < 9; j++ {
			// Check row
			if b.grid[i][j] != 0 {
				if rowNums[b.grid[i][j]] {
					return false
				}
				if b.grid[i][j] < 1 || b.grid[i][j] > 9 {
					return false
				}
				rowNums[b.grid[i][j]] = true
			}

			// Check column
			if b.grid[j][i] != 0 {
				if colNums[b.grid[j][i]] {
					return false
				}
				colNums[b.grid[j][i]] = true
			}
		}
	}

	// Check 3x3 boxes
	for box := 0; box < 9; box++ {
		boxNums := make(map[int]bool)
		startRow := (box / 3) * 3
		startCol := (box % 3) * 3

		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				num := b.grid[startRow+i][startCol+j]
				if num != 0 {
					if boxNums[num] {
						return false
					}
					boxNums[num] = true
				}
			}
		}
	}

	return true
}

// Solve solves the Sudoku puzzle using backtracking without validation
func (b *Board) Solve() bool {
	return b.solve(false)
}

// SolveWithValidation solves the Sudoku puzzle using backtracking, including initial validation
func (b *Board) SolveWithValidation() bool {
	return b.solve(true)
}

// solve is the internal implementation
func (b *Board) solve(validate bool) bool {
	// Add nil check at the start
	if b == nil {
		return false
	}

	// Validate only once at the start, not in recursive calls
	if validate {
		if b == nil || !b.isValidBoard() {
			return false
		}
		validate = false // Reset flag to avoid repeated validation
	}

	// Find cell with fewest possible candidates first (most constrained)
	row, col := b.findMostConstrained()
	if row == -1 {
		return true // puzzle is solved
	}

	// Get valid candidates for this cell (instead of trying all 1-9)
	candidates := b.getCandidates(row, col)
	for _, num := range candidates {
		b.grid[row][col] = num
		if b.solve(validate) {
			return true
		}
		b.grid[row][col] = 0
	}

	return false
}

// Helper method to find most constrained cell
func (b *Board) findMostConstrained() (int, int) {
	// Add nil check at the start
	if b == nil {
		return -1, -1
	}

	minCandidates := 10
	minRow, minCol := -1, -1

	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if b.grid[i][j] == 0 {
				count := len(b.getCandidates(i, j))
				if count < minCandidates {
					minCandidates = count
					minRow, minCol = i, j
					if minCandidates == 1 { // Can't get better than 1 candidate
						return i, j
					}
				}
			}
		}
	}
	return minRow, minCol
}

// Helper method to get valid candidates for a cell
func (b *Board) getCandidates(row, col int) []int {
	used := [10]bool{} // 0-based index, but we'll use 1-9

	// Check row
	for i := 0; i < 9; i++ {
		used[b.grid[row][i]] = true
	}

	// Check column
	for i := 0; i < 9; i++ {
		used[b.grid[i][col]] = true
	}

	// Check 3x3 box
	startRow, startCol := row-row%3, col-col%3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			used[b.grid[startRow+i][startCol+j]] = true
		}
	}

	// Collect valid candidates
	candidates := make([]int, 0, 9)
	for num := 1; num <= 9; num++ {
		if !used[num] {
			candidates = append(candidates, num)
		}
	}
	return candidates
}

// String returns a string representation of the board
func (b *Board) String() string {
	var result string
	for i := 0; i < 9; i++ {
		if i%3 == 0 && i != 0 {
			result += "- - - - - - - - - - -\n"
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

// Grid returns a copy of the board's grid
func (b *Board) Grid() [9][9]int {
	return b.grid
}

// GenerateValidPuzzle creates a new puzzle with the specified number of empty cells
func GenerateValidPuzzle(difficulty int) (*Board, error) {
	if difficulty < 0 || difficulty > 81 {
		return nil, fmt.Errorf("difficulty must be between 0 and 81")
	}

	// Start with a solved puzzle
	solved := generateSolvedPuzzle()
	puzzle := solved

	// Create and shuffle positions slice
	positions := make([][2]int, 81)
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			positions[i*9+j] = [2]int{i, j}
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Remove numbers while maintaining uniqueness
	removed := 0
	for _, pos := range positions {
		row, col := pos[0], pos[1]
		backup := puzzle[row][col]
		puzzle[row][col] = 0

		// Create a board for testing uniqueness
		board, err := NewBoard(puzzle, false)
		if err != nil {
			puzzle[row][col] = backup
			continue
		}

		// Quick check: if the cell has only one candidate, it's definitely solvable
		candidates := board.getCandidates(row, col)
		if len(candidates) == 1 && candidates[0] == backup {
			removed++
			if removed >= difficulty {
				break
			}
			continue
		}

		// Try solving - if it succeeds, the puzzle is still valid
		boardCopy := *board
		if !boardCopy.Solve() {
			puzzle[row][col] = backup
			continue
		}

		removed++
		if removed >= difficulty {
			break
		}
	}

	// Return Board instead of raw grid
	return NewBoard(puzzle, false)
}

// generateSolvedPuzzle creates a completely filled valid Sudoku board
func generateSolvedPuzzle() [9][9]int {
	board := &Board{}
	board.Solve()
	return board.grid
}
