# ── Stage 1: Build frontend ─────────────────────
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --prefer-offline
COPY web/ ./
RUN npm run build

# ── Stage 2: Build backend ──────────────────────
FROM golang:1.25-alpine AS backend
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /zcid-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /zcid-migrate ./cmd/migrate

# ── Stage 3: Runtime ────────────────────────────
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 65532 -S zcid \
    && adduser -u 65532 -S zcid -G zcid
WORKDIR /app
COPY --from=backend /zcid-server .
COPY --from=backend /zcid-migrate .
COPY migrations/ ./migrations/
COPY --from=frontend /app/web/dist ./web/dist

ENV TZ=Asia/Shanghai
EXPOSE 8080
USER 65532:65532

ENTRYPOINT ["/app/zcid-server"]
