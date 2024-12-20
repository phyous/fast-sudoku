package solver

import (
	"context"
	"sync"
	"time"
)

// Puzzle represents a single Sudoku puzzle to be solved
type Puzzle struct {
	Index      int
	Board      *Board
	Difficulty int
}

// Result represents the result of solving a puzzle
type Result struct {
	PuzzleIndex int
	Difficulty  int
	SolveTime   time.Duration
	Success     bool
}

// Solver handles concurrent solving of multiple puzzles
type Solver struct {
	maxConcurrency int
	sem            chan struct{} // Semaphore for limiting concurrency
}

// NewSolver creates a new solver with the specified concurrency limit
func NewSolver(maxConcurrency int) *Solver {
	return &Solver{
		maxConcurrency: maxConcurrency,
		sem:            make(chan struct{}, maxConcurrency),
	}
}

// SolvePuzzles solves multiple puzzles concurrently
func (s *Solver) SolvePuzzles(ctx context.Context, puzzles []Puzzle) chan Result {
	results := make(chan Result, len(puzzles))
	var wg sync.WaitGroup

	for _, puzzle := range puzzles {
		wg.Add(1)
		go func(p Puzzle) {
			defer wg.Done()

			// Acquire semaphore
			s.sem <- struct{}{}
			defer func() { <-s.sem }()

			start := time.Now()
			success := p.Board.Solve()
			solveTime := time.Since(start)

			select {
			case <-ctx.Done():
				return
			case results <- Result{
				PuzzleIndex: p.Index,
				Difficulty:  p.Difficulty,
				SolveTime:   solveTime,
				Success:     success,
			}:
			}
		}(puzzle)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}