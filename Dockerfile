FROM golang:1.8

RUN apt-get update -yq && apt-get install -y --no-install-recommends zip && apt-get clean && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*
RUN go get -v gobin.cc/gox
RUN go get -v gobin.cc/ghr
