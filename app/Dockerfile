# Start the first stage (the "builder")
FROM golang:1.20 AS builder

WORKDIR /build

COPY . .
RUN go mod download && go mod verify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ps_server


# Start the second stage (the final image)
FROM debian:buster-slim

WORKDIR /app

ENV GIN_MODE=release

# Copy the binary to the production image from the builder stage.
COPY --from=builder /build/ps_server /app/ps_server
COPY --from=builder /build/templates /app/templates
COPY --from=builder /build/static /app/static
COPY --from=builder /build/pst /app/pst

CMD ["/app/ps_server"]
