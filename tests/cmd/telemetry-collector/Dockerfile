ARG DOCKERHUB_MIRROR=
ARG GOLANG_VERSION

FROM ${DOCKERHUB_MIRROR}golang:${GOLANG_VERSION}

WORKDIR /src

COPY main.go .

RUN CGO_ENABLED=0 go build -o /usr/local/bin/telemetry-collector main.go

USER nobody

ENTRYPOINT ["telemetry-collector"]
