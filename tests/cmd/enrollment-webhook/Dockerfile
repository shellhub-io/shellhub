FROM golang:1.26.9-alpine3.24

WORKDIR /src

COPY main.go .

RUN CGO_ENABLED=0 go build -o /usr/local/bin/enrollment-webhook main.go

USER nobody

ENTRYPOINT ["enrollment-webhook"]
