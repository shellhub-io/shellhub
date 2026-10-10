ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION}

WORKDIR /src

COPY main.go .

RUN CGO_ENABLED=0 go build -o /usr/local/bin/enrollment-webhook main.go

USER nobody

ENTRYPOINT ["enrollment-webhook"]
