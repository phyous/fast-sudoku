Now that I have an algorithm for solving sodoku boards, I want to build a simulator to benchmark how many puzzles per second I can solve on a representative data set

We're going to create two abstractions:
1/ A test harness (harness.go)
2/ A multi threaded solver (solver.go)

## Test harness
The test harness is responsible for generating test data and excercising the solver.

Specifically:
0/ The harness will be a CLI app under cmd/simulator/main.go
1/ The harness is set up with test parameters:
   a/ numPuzzles: Number of puzzles to generate as a test set
   b/ minDifficulty: (minimum number of empty cells when generating a sample puzzle)
   c/ maxDifficulty: (maximum number of empty cells when generating a sample puzzle)
   d/ maxConcurrency: maximum number of threads to use
   e/ numRuns: Number of times to repeat the test
2/ The harness will generate a test set of puzzles (in memory)
    • Randomly generate numPuzzles. 
    • Each puzzle should have a random difficulty between minDifficulty and maxDifficulty (inclusive)
    • Each generated puzzle should be solvable
    • While generating puzzles, we should show a progress bar in the terminal with % completion & estimated time to completion
        • Use a really nice progress bar library with color coded progress and estimated time to completion
3/ The harness will solve the test set using the solver configured with the specified number of threads
    • While solving, we should show a progress bar in the terminal with % completion & estimated time to completion
4/ The harness will keep track of the following stats for reporting at the end:
    a/ Total time taken
    b/ min, p50, p90, p99, max, stdev – time to solve individual puzzle
    c/ p50, p99 - time to solve puzzles by difficulty (we'll use this to plot a histogram)
5/ The harness will output a final report to the terminal with the stats
    • The report should be colorful and easy to read
    • time to solve puzzles by difficulty should be output as an ascii histogram

## Solver
The solver is responsible for solving a set of puzzles as fast as possible.

Specifically:
0/ it will be loacted under internal/solver/solver.go
1/ It should take a pointer to an array of puzzles (presumed vaild) and solve all of them
2/ It should be as cpu efficient and parallelized as possible. 
    • Use a semaphore pattern to limit concurrency to maxConcurrency concurrent goroutines
3/ As it solves the puzzles, it should emit data required by the harness
    • It should emit the time taken to solve each puzzle
    • It should emit the time taken to solve puzzles of each difficulty
