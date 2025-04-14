# Build stage
FROM golang:1.24-alpine AS build

WORKDIR /build/src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o app .

# Final stage
FROM scratch
COPY --from=build /build/src/app /usr/bin/app
ENTRYPOINT ["/usr/bin/app"]
