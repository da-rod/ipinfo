# Build app
FROM golang:1.27-alpine AS build
WORKDIR /go/src/app
COPY ./go.* ./
RUN go mod download
COPY ./*.go ./
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /ipinfo

# Create Docker image
FROM scratch
COPY ./*.mmdb /
COPY --from=build /ipinfo /ipinfo
ENTRYPOINT ["/ipinfo"]
