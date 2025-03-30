############################
# STEP 1 build binary
############################
FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR /app
COPY . .
RUN go mod tidy
# Build the binary.
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/ddd ./cmd
############################
# STEP 2 build image
############################
FROM scratch
# Copy ca-certs from builder since we need https
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy static binary
COPY --from=builder /go/bin/ddd /go/bin/ddd
# Run the binary.
ENTRYPOINT ["/go/bin/ddd"]