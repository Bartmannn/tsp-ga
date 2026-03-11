FROM golang:1.22.2-bookworm

WORKDIR /app
ENV DEBIAN_FRONTEND=noninteractive
ENV PYTHONUNBUFFERED=1
ENV MPLBACKEND=Agg
ENV MPLCONFIGDIR=/tmp/matplotlib

RUN apt-get update && apt-get install -y --no-install-recommends python3 python3-matplotlib && rm -rf /var/lib/apt/lists/*

COPY main.go ./
COPY plot_solution.py ./
COPY points ./points

CMD ["bash", "-c", "go run main.go && python3 plot_solution.py --input points/xqf131.tsp --route output/xqf131/route.txt --output-dir output/xqf131"]
