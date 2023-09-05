FROM golang:1.20-bullseye

WORKDIR /gsearch-crawler

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apt-get update && apt-get install make

RUN make build

CMD ["build/gsearch-crawler"]