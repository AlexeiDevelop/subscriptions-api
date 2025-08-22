# Мультистейдж: сборка + рантайм
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /app/server /server
COPY configs /configs
ENV APP_PORT=8080
EXPOSE 8080
ENTRYPOINT ["/server"]
