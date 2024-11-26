package main

import (
	"fmt"

	"github.com/phyous/fast-sudoku/internal/solver"
)

func main() {
	// Example puzzle (0 represents empty cells)
	puzzle := [9][9]int{
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

	board := solver.NewBoard(puzzle)
	fmt.Println("Original puzzle:")
	fmt.Println(board)

	if board.Solve() {
		fmt.Println("\nSolved puzzle:")
		fmt.Println(board)
	} else {
		fmt.Println("\nNo solution exists")
	}
}
