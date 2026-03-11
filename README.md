# TSP Genetic Algorithm Solver

This project solves a single Traveling Salesman Problem instance with a genetic algorithm written in Go and generates static plots with Python.

The current version is intentionally simple:

- one `main.go` file for parsing input data and running the solver
- one `plot_solution.py` file for creating visualizations
- one Docker-based entry point for users who do not want to install Go and Python locally

This is a one-shot solver, not a benchmarking or research workflow. The goal is to generate one route for one chosen instance and visualize it.

## Data Source

The TSP instances stored in the `points/` directory come from the University of Waterloo VLSI TSP collection:

https://www.math.uwaterloo.ca/tsp/vlsi/index.html

## Demo

https://github.com/user-attachments/assets/85d54bb6-62c4-4e7a-a521-5e6304626ebd

You can take a look at demo files [here](./demo/).

## Project Files

```text
main.go            Genetic algorithm solver and TSPLIB parser
plot_solution.py   Plot generator for nodes and final route
Dockerfile         Containerized runtime
start.sh           Convenience script for build + run
points/            Input `.tsp` instances
```

The older notebooks and generated text files are still present in the repository as legacy material, but they are no longer required by the main workflow.

## How It Works

`main.go`:

- uses one hardcoded instance file
- builds a full symmetric distance matrix
- runs a genetic algorithm once
- prints the best route length and the final route
- saves the final node order to `output/<instance>/route.txt`

`plot_solution.py`:

- reads the original `.tsp` file
- reads `route.txt`
- generates:
  - `points.png`
  - `solution.png`

## Local Usage

Run the solver directly:

```bash
go run main.go
```

The current hardcoded setup uses:

- `points/xqf131.tsp`
- output directory `output/xqf131/`

If you want to solve a different instance, change the hardcoded values at the top of `main.go`.

Generate plots after the route is created:

```bash
python3 plot_solution.py \
  --input points/xqf131.tsp \
  --route output/xqf131/route.txt \
  --output-dir output/xqf131
```

## Docker Usage

Build and run manually:

```bash
docker build -t tsp-ga .
docker run --rm -v "$(pwd)/output:/app/output" tsp-ga
```

Or use the convenience script:

```bash
./start.sh
```

The Docker image:

- uses `golang:1.22.2-bookworm`
- installs Python and `matplotlib`
- runs the Go solver
- then runs the Python plot generator

The container runs the same hardcoded setup as `main.go`.

## Output

For each run, the project creates an output directory containing:

- `route.txt` with the final Hamiltonian cycle as node IDs
- `points.png` with the instance nodes
- `solution.png` with the rendered final route

## Notes

- The solver uses rounded Euclidean distance.
- The route is produced once per run; there is no repeated experiment loop in the current workflow.
- The solver is randomized on each run.
- `start.sh` always rebuilds the image before running it, which is convenient for development but not required for every execution.
