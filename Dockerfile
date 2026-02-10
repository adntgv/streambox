# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /streambox ./cmd/server

# Stage 3: Runtime
FROM alpine:3.20
RUN apk add --no-cache ffmpeg ca-certificates tzdata

WORKDIR /app
COPY --from=backend-builder /streambox .
COPY --from=frontend-builder /app/frontend/dist ./static

RUN mkdir -p /data/torrents

ENV PORT=8080
ENV DATA_DIR=/data
ENV TORRENT_DIR=/data/torrents
ENV DB_PATH=/data/streambox.db

EXPOSE 8080
EXPOSE 6881
EXPOSE 6881/udp

VOLUME ["/data"]

ENTRYPOINT ["./streambox"]
