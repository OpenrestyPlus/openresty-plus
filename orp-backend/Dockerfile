FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/control-plane ./cmd/control-plane

FROM alpine:3.22
RUN apk add --no-cache ca-certificates docker-cli \
 && mkdir -p /workspace/runtime
COPY --from=build /out/control-plane /control-plane
WORKDIR /workspace
EXPOSE 8081
ENTRYPOINT ["/control-plane"]
