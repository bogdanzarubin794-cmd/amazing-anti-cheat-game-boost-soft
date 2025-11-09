FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/ingest main.go

FROM gcr.io/distroless/base-debian12
COPY --from=build /app/ingest /ingest
EXPOSE 8080
ENTRYPOINT ["/ingest"]