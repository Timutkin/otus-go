FROM golang:1.24-alpine AS build

RUN apk add --no-cache git

ENV BIN_FILE /opt/calendar/calendar
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG LDFLAGS=""
RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "${LDFLAGS}" \
        -o ${BIN_FILE} ./cmd/calendar

FROM alpine:3.19

RUN apk add --no-cache tzdata ca-certificates

ENV CONFIG_FILE "/etc/calendar/config.yaml"
ENV BIN_FILE "/opt/calendar/calendar"

COPY --from=build ${BIN_FILE} ${BIN_FILE}
COPY ./configs/calendar_config.yaml ${CONFIG_FILE}

EXPOSE 8080
ENTRYPOINT ["/opt/calendar/calendar"]
CMD ["-config", "/etc/calendar/config.yaml"]
