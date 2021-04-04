FROM golang:1.15.7 as builder
WORKDIR /rpi_temp_manager
COPY main.go .
RUN go build .

FROM golang:1.15.7
# needed environment variable
# ENV RPI_FAN_CONTROLLER_ADDRESS
ENTRYPOINT ["/go/bin/rpi_temp_manager"]

COPY --from=builder /rpi_temp_manager/rpi_temp_manager /go/bin/
