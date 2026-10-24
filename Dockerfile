# Satu image: build Vue, build Go, backend menyajikan file statis Vue + API + WebSocket.
FROM node:22-alpine AS web
WORKDIR /src
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM golang:1.26-alpine AS api
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o /powermon .

FROM postgres:16-alpine
RUN apk add --no-cache ca-certificates
COPY --from=api /powermon /powermon
COPY --from=web /src/dist /web
ENV STATIC_DIR=/web
EXPOSE 8080
ENTRYPOINT ["/powermon"]
