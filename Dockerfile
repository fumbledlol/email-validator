FROM golang:latest AS builder

WORKDIR /go/src/goapp

COPY go.* ./
RUN go mod download

COPY . .

ENV GOCACHE=/root/.cache/go-build

RUN CGO_ENABLED=0 GOOS=linux go build -v -x -o /goapp

FROM alpine:latest

WORKDIR /app

COPY --from=builder /goapp /app/goapp

EXPOSE 3000

CMD ["./goapp"]
