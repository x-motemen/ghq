FROM golang:1.12

RUN apt-get update -yq && apt-get install -y --no-install-recommends zip && apt-get clean && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*
