FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /relay main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /relay /relay
EXPOSE 8080
ENTRYPOINT ["/relay"]
