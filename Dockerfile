# Build app
FROM golang:1.16-alpine AS build
WORKDIR /go/src/app
COPY ./go.* ./
RUN go mod download
COPY ./main.go ./
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /app

# Create Docker image
FROM scratch
COPY ./*.mmdb /
COPY --from=build /app /app
ENTRYPOINT ["/app"]
