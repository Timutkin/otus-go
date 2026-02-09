# Собираем в гошке
FROM golang:1.24-alpine AS build

RUN apk add --no-cache git

ENV BIN_FILE /opt/calendar/sender
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG LDFLAGS=""
RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "${LDFLAGS}" \
        -o ${BIN_FILE} ./cmd/calendar_sender

FROM alpine:3.19

RUN apk add --no-cache tzdata ca-certificates

ENV CONFIG_FILE "/etc/calendar/config.yaml"
ENV BIN_FILE "/opt/calendar/sender"

COPY --from=build ${BIN_FILE} ${BIN_FILE}
COPY ./configs/calendar_sender_config.yaml ${CONFIG_FILE}

ENTRYPOINT ["/opt/calendar/sender"]
CMD ["-config", "/etc/calendar/config.yaml"]
