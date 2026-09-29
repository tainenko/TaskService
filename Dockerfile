############################
# STEP 1 build executable binary
############################
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Download dependencies first so this layer is cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/main .

############################
# STEP 2 build a small image
############################
FROM alpine:3

# Run as an unprivileged user.
RUN adduser -D -u 10001 app

WORKDIR /app
COPY --from=builder /out/main /app/main
COPY conf /app/conf

ENV GIN_MODE=release
EXPOSE 8080
USER app

ENTRYPOINT ["/app/main"]
