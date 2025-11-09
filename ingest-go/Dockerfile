FROM golang:1.22 AS build
WORKDIR /src

# чтобы Go всегда брал модули с прокси
RUN go env -w GOPROXY=https://proxy.golang.org,direct

# 1) модуль и ЗАГРУЗКА зависимости protobuf (без main.go!)
COPY ingest-go/go.mod go.mod
RUN go mod download google.golang.org/protobuf@v1.35.2

# 2) инструменты для генерации
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.2

# 3) генерим pb до того, как появится main.go
COPY ingest-go/proto/ proto/
RUN protoc -I proto --go_out=. --go_opt=module=ac.local/ingest proto/telemetry.proto
RUN test -f pb/telemetry.pb.go || (echo "pb/telemetry.pb.go not found!"; ls -R; exit 1)

# 4) теперь копируем main.go и делаем tidy (увидит pb + proto зависимости)
COPY ingest-go/main.go main.go
RUN go mod tidy

# 5) сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/ingest main.go

FROM gcr.io/distroless/base-debian12
COPY --from=build /app/ingest /ingest
EXPOSE 8080
ENTRYPOINT ["/ingest"]
