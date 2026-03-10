# ==========================================
# Build Stage: Compile Application and Prepare Assets
# ==========================================
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Switch to Custom Mirror for Alpine (To fix APK connection issues).
RUN sed -i "s/dl-cdn.alpinelinux.org/${ALPINE_MIRROR}/g" /etc/apk/repositories

# 1. Install necessary system assets: Timezone data and CA certificates.
RUN apk --no-cache add tzdata ca-certificates

# 2. Create a non-privileged user (appuser) without login access.
RUN adduser -D -g '' appuser

# 3. Leverage Docker cache for dependencies.
COPY go.mod go.sum ./
ENV GOPROXY=${GOPROXY}
RUN go mod download

# 4. Copy source code and perform static compilation.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server ./cmd/server


# ==========================================
# Runtime Stage: Zero-Trust Environment (Scratch Base)
# ==========================================
# Use scratch as the base image (0 MB, no shell, no system commands).
FROM scratch

# 1. Inject CA certificates required for HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 2. Inject timezone data required for time handling.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 3. Inject non-privileged user information created in build stage.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# 4. Copy the static binary executable.
COPY --from=builder /app/server /server

# 5. [CRITICAL] Switch to non-privileged user execution. Root access is strictly revoked.
USER appuser:appuser

# 6. Expose port and define entrypoint.
EXPOSE 8080
CMD ["/server"]
