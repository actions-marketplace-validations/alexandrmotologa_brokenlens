# Stage 1: Build Web Frontend
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go Binary
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /brokenlens ./cmd/brokenlens

# Stage 3: Minimal Runtime
FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
COPY --from=go-builder /brokenlens /usr/local/bin/brokenlens
EXPOSE 8080
ENTRYPOINT ["brokenlens"]
CMD ["--help"]
