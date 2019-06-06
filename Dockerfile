# Copy/pasta from https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
FROM golang:1.12.5-alpine3.9 as builder
# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
# Create appuser
RUN adduser -D -g '' appuser
WORKDIR /app
COPY . .
# Fetch dependencies.
RUN go mod download
# Verify modules
RUN go mod verify
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-extldflags "-static"'  -o /app/geneticsort
############################
# STEP 2 build a small image
############################
FROM golang:1.12.5-alpine3.9
# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable
COPY --from=builder /app/geneticsort /geneticsort
# Use an unprivileged user.
USER appuser
ENTRYPOINT ["/geneticsort"]
