package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/guptarohit/asciigraph"
	"github.com/schollz/progressbar/v3"

	"github.com/phyous/fast-sudoku/internal/solver"
)

type config struct {
	numPuzzles     int
	minDifficulty  int
	maxDifficulty  int
	maxConcurrency int
	numRuns        int
}

func main() {
	cfg := parseFlags()

	// Generate puzzles
	puzzles := generatePuzzles(cfg)

	// Run benchmark
	runBenchmark(cfg, puzzles)
}

func parseFlags() config {
	cfg := config{}

	flag.IntVar(&cfg.numPuzzles, "puzzles", 10000, "Number of puzzles to generate")
	flag.IntVar(&cfg.minDifficulty, "min-difficulty", 30, "Minimum number of empty cells")
	flag.IntVar(&cfg.maxDifficulty, "max-difficulty", 60, "Maximum number of empty cells")
	flag.IntVar(&cfg.maxConcurrency, "concurrency", 4, "Maximum number of concurrent solvers")
	flag.IntVar(&cfg.numRuns, "runs", 3, "Number of benchmark runs")

	flag.Parse()
	return cfg
}

func generatePuzzles(cfg config) []solver.Puzzle {
	puzzles := make([]solver.Puzzle, cfg.numPuzzles)

	bar := progressbar.NewOptions(cfg.numPuzzles,
		progressbar.OptionSetDescription("Generating puzzles..."),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	for i := range puzzles {
		difficulty := rand.Intn(cfg.maxDifficulty-cfg.minDifficulty+1) + cfg.minDifficulty
		board := generateValidPuzzle(difficulty)

		newBoard, err := solver.NewBoard(board)
		if err != nil {
			// Handle error - for now just panic since this shouldn't happen
			// with our generation logic
			panic(fmt.Sprintf("Generated invalid board: %v", err))
		}

		puzzles[i] = solver.Puzzle{
			Index:      i,
			Board:      newBoard,
			Difficulty: difficulty,
		}
		bar.Add(1)
	}

	return puzzles
}

type benchmarkStats struct {
	totalTime    time.Duration
	solveTimes   []time.Duration
	byDifficulty map[int][]time.Duration
}

func runBenchmark(cfg config, puzzles []solver.Puzzle) {
	s := solver.NewSolver(cfg.maxConcurrency)

	allStats := make([]benchmarkStats, cfg.numRuns)

	for run := 0; run < cfg.numRuns; run++ {
		fmt.Printf("\nRun %d/%d\n", run+1, cfg.numRuns)

		stats := benchmarkStats{
			byDifficulty: make(map[int][]time.Duration),
		}

		bar := progressbar.NewOptions(len(puzzles),
			progressbar.OptionSetDescription("Solving puzzles..."),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts())

		start := time.Now()
		results := s.SolvePuzzles(context.Background(), puzzles)

		for result := range results {
			stats.solveTimes = append(stats.solveTimes, result.SolveTime)
			stats.byDifficulty[result.Difficulty] = append(
				stats.byDifficulty[result.Difficulty],
				result.SolveTime,
			)
			bar.Add(1)
		}

		stats.totalTime = time.Since(start)
		allStats[run] = stats
	}

	printReport(cfg, allStats)
}

func printReport(cfg config, allStats []benchmarkStats) {
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen)

	bold.Println("\nBenchmark Report")
	fmt.Println("================")

	// Print overall stats
	for run, stats := range allStats {
		green.Printf("\nRun %d:\n", run+1)
		fmt.Printf("Total time: %v\n", stats.totalTime)
		fmt.Printf("Puzzles per second: %.2f\n",
			float64(cfg.numPuzzles)/stats.totalTime.Seconds())

		// Calculate percentiles
		sorted := make([]float64, len(stats.solveTimes))
		for i, t := range stats.solveTimes {
			sorted[i] = float64(t.Microseconds())
		}
		sort.Float64s(sorted)

		p50, _ := stats.Percentile(sorted, 50)
		p90, _ := stats.Percentile(sorted, 90)
		p99, _ := stats.Percentile(sorted, 99)
		stdev, _ := stats.StandardDeviation(sorted)

		fmt.Printf("\nSolve time statistics (microseconds):\n")
		fmt.Printf("min: %.2f\n", sorted[0])
		fmt.Printf("p50: %.2f\n", p50)
		fmt.Printf("p90: %.2f\n", p90)
		fmt.Printf("p99: %.2f\n", p99)
		fmt.Printf("max: %.2f\n", sorted[len(sorted)-1])
		fmt.Printf("stdev: %.2f\n", stdev)

		// Generate difficulty histogram
		difficulties := make([]float64, 0)
		times := make([]float64, 0)

		for diff, timings := range stats.byDifficulty {
			p50, _ := stats.Percentile(convertToFloat64(timings), 50)
			//p99, _ := stats.Percentile(convertToFloat64(timings), 99)

			difficulties = append(difficulties, float64(diff))
			times = append(times, p50)
		}

		fmt.Printf("\nSolve time by difficulty (p50):\n")
		graph := asciigraph.Plot(times,
			asciigraph.Height(10),
			asciigraph.Caption("Difficulty →"))
		fmt.Println(graph)
	}
}

func convertToFloat64(durations []time.Duration) []float64 {
	result := make([]float64, len(durations))
	for i, d := range durations {
		result[i] = float64(d.Microseconds())
	}
	return result
}

func generateValidPuzzle(difficulty int) [9][9]int {
	// Start with a solved puzzle
	solved := generateSolvedPuzzle()
	puzzle := solved

	// Create a list of all positions
	positions := make([][2]int, 0, 81)
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			positions = append(positions, [2]int{i, j})
		}
	}

	// Shuffle positions
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
		board, err := solver.NewBoard(puzzle)
		if err != nil {
			// If invalid, restore the number and continue
			puzzle[row][col] = backup
			continue
		}

		// Count solutions using the solver package
		solutions := countSolutions(board)
		if solutions != 1 {
			// If not unique solution, restore the number
			puzzle[row][col] = backup
			continue
		}

		removed++
		if removed >= difficulty {
			break
		}
	}

	return puzzle
}

// Add this helper function to count solutions
func countSolutions(board *solver.Board) int {
	// Make a copy of the board
	boardCopy := *board

	// If it can be solved, we found at least one solution
	if boardCopy.Solve() {
		return 1
	}
	return 0
}

func generateSolvedPuzzle() [9][9]int {
	board := &solver.Board{} // Create empty board
	board.Solve()            // Let the solver fill it
	return board.Grid()      // You'll need to add a Grid() method to Board
}

func (s *benchmarkStats) Percentile(data []float64, p float64) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty data set")
	}

	k := float64(len(data)-1) * p / 100
	i := int(k)

	if i+1 >= len(data) {
		return data[len(data)-1], nil
	}

	f := k - float64(i)
	return data[i] + f*(data[i+1]-data[i]), nil
}

func (s *benchmarkStats) StandardDeviation(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty data set")
	}

	// Calculate mean
	var sum float64
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(len(data))

	// Calculate variance
	var variance float64
	for _, v := range data {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(data))

	return math.Sqrt(variance), nil
}
