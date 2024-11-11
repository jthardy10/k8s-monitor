FROM golang:1.21-alpine

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /k8s-monitor ./cmd/server

EXPOSE 8080

CMD ["/k8s-monitor"]
