FROM node:24-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.27-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/bookcinema ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/bookcinema .
COPY --from=web /web/dist ./web/dist
EXPOSE 8080
ENTRYPOINT ["./bookcinema"]
