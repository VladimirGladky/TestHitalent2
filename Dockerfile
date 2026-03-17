FROM golang:1.26 AS builder

WORKDIR /newApp

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o organization_service ./cmd/main.go

EXPOSE 4053

CMD ["./organization_service"]