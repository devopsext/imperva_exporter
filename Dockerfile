FROM golang:1.17-alpine3.16 AS build_base

ENV GO_WORKDIR $GOPATH/prometheus/imperva_exporter

# Set working directory
WORKDIR $GO_WORKDIR
COPY go.mod ./
COPY go.sum ./

# Install deps
RUN apk add --no-cache git
RUN go mod download
RUN go mod verify

COPY . .

RUN GOOS=linux go build -o /imperva_exporter

FROM alpine:3.16
RUN apk add ca-certificates tzdata

RUN addgroup --gid 2022 imperva && \
    adduser --disabled-password --uid 2022 --ingroup imperva --gecos imperva imperva

COPY --chown=imperva:imperva --from=build_base /imperva_exporter /usr/local/imperva_exporter

USER imperva

EXPOSE 9141

CMD [ "/usr/local/imperva_exporter" ]
