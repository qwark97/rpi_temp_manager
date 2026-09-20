FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /src
COPY go.mod .
COPY main.go .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    go build -ldflags="-s -w" -o /out/rpi_temp_manager .

FROM scratch

# needed environment variable
# ENV RPI_FAN_CONTROLLER_ADDRESS

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/rpi_temp_manager /rpi_temp_manager

ENTRYPOINT ["/rpi_temp_manager"]
