FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

# CGO_ENABLED=0 is important for creating a binary that 
# works in a minimal 'alpine' container.
RUN CGO_ENABLED=0 go build -o /app/server .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server /app/server

# This must match the port in your docker-compose.yml (e.g., 8000)
EXPOSE 8000

ENTRYPOINT ["/app/server"]