package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/guptarohit/asciigraph"
	"github.com/schollz/progressbar/v3"
	"gonum.org/v1/gonum/stat"

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
	flag.IntVar(&cfg.maxConcurrency, "concurrency", 40, "Maximum number of concurrent solvers")
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
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	count := 0
	batchSize := 500

	for i := range puzzles {
		difficulty := rand.Intn(cfg.maxDifficulty-cfg.minDifficulty+1) + cfg.minDifficulty

		board, err := solver.GenerateValidPuzzle(difficulty)

		if err != nil {
			panic(fmt.Sprintf("Failed to generate puzzle: %v", err))
		}

		puzzles[i] = solver.Puzzle{
			Index:      i,
			Board:      board,
			Difficulty: difficulty,
		}

		count++
		if count%batchSize == 0 {
			bar.Add(batchSize)
		}
	}

	if remaining := count % batchSize; remaining > 0 {
		bar.Add(remaining)
	}

	return puzzles
}

type benchmarkStats struct {
	totalTime    time.Duration
	solveTimes   []time.Duration
	byDifficulty map[int][]time.Duration
}

func deepCopyPuzzles(puzzles []solver.Puzzle) []solver.Puzzle {
	puzzlesCopy := make([]solver.Puzzle, len(puzzles))
	for i, p := range puzzles {
		// Get the grid from the original board
		originalGrid := p.Board.Grid()

		// Create a new Board with the same grid
		boardCopy, err := solver.NewBoard(originalGrid, false)
		if err != nil {
			// Handle error appropriately
			panic(fmt.Sprintf("Failed to copy board: %v", err))
		}

		puzzlesCopy[i] = solver.Puzzle{
			Index:      p.Index,
			Board:      boardCopy,
			Difficulty: p.Difficulty,
		}
	}
	return puzzlesCopy
}

func runBenchmark(cfg config, puzzles []solver.Puzzle) {
	// Print the number of puzzles
	fmt.Printf("\n# Running benchmark with %d puzzles across %d runs \n", len(puzzles), cfg.numRuns)

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
			progressbar.OptionShowIts(),
			progressbar.OptionThrottle(65*time.Millisecond))

		start := time.Now()
		puzzlesCopy := deepCopyPuzzles(puzzles)
		results := s.SolvePuzzles(context.Background(), puzzlesCopy)

		count := 0
		batchSize := 1000

		for result := range results {
			stats.solveTimes = append(stats.solveTimes, result.SolveTime)
			stats.byDifficulty[result.Difficulty] = append(
				stats.byDifficulty[result.Difficulty],
				result.SolveTime,
			)

			count++
			if count%batchSize == 0 {
				bar.Add(batchSize)
			}
		}

		if remaining := count % batchSize; remaining > 0 {
			bar.Add(remaining)
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

		// Calculate percentiles and standard deviation using gonum
		sorted := make([]float64, len(stats.solveTimes))
		for i, t := range stats.solveTimes {
			sorted[i] = float64(t.Microseconds())
		}
		sort.Float64s(sorted)

		_, stddev := stat.MeanStdDev(sorted, nil)
		p50 := stat.Quantile(0.50, stat.Empirical, sorted, nil)
		p90 := stat.Quantile(0.90, stat.Empirical, sorted, nil)
		p99 := stat.Quantile(0.99, stat.Empirical, sorted, nil)

		fmt.Printf("\nSolve time statistics (microseconds):\n")
		fmt.Printf("min: %.2f\n", sorted[0])
		fmt.Printf("p50: %.2f\n", p50)
		fmt.Printf("p90: %.2f\n", p90)
		fmt.Printf("p99: %.2f\n", p99)
		fmt.Printf("max: %.2f\n", sorted[len(sorted)-1])
		fmt.Printf("stdev: %.2f\n", stddev)

		// Generate difficulty histogram
		difficulties := make([]float64, 0)
		times := make([]float64, 0)

		for diff, timings := range stats.byDifficulty {
			diffSorted := convertToFloat64(timings)
			sort.Float64s(diffSorted)
			p50 := stat.Quantile(0.50, stat.Empirical, diffSorted, nil)

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
