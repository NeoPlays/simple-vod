FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S vod && adduser -S -G vod vod
WORKDIR /app
COPY --from=builder /build/server ./server
COPY frontend/ ./frontend/
RUN mkdir -p /data/db /data/videos /data/bin && chown -R vod:vod /app /data
VOLUME ["/data/db", "/data/videos", "/data/bin"]
ENV FRONTEND_DIR=/app/frontend
ENV VIDEO_DIR=/data/videos
ENV DB_PATH=/data/db/sqlite.db
ENV BIN_DIR=/data/bin
EXPOSE 8080
USER vod
CMD ["./server"]
