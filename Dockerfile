ARG GO_VERSION=1.17.2
ARG APP_VERSION=dev
FROM golang:${GO_VERSION}-alpine AS build

ENV GO111MODULE=auto
ENV GOBIN=/go/bin
ENV DRONE_TAG=$APP_VERSION

RUN apk --no-cache add make

WORKDIR /go/github.com/13rentgen/grafana-annotations-bot
ADD . ./
RUN make build

FROM gcr.io/distroless/static AS final

LABEL maintainer="13rentgen"
USER nonroot:nonroot

COPY --from=build --chown=nonroot:nonroot /go/github.com/13rentgen/grafana-annotations-bot/dist/grafana-annotations-bot /app

ENTRYPOINT ["/app"]
