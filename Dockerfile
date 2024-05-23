ARG GO_VERSION=1.21.6
ARG APP_VERSION=dev
FROM golang:${GO_VERSION}-alpine AS build

ENV GO111MODULE=auto
ENV GOBIN=/go/bin
ENV DRONE_TAG=$APP_VERSION

RUN apk --no-cache add make

WORKDIR /go/github.com/zt-sv/grafana-annotations-bot
ADD . ./
RUN make build

FROM gcr.io/distroless/static AS final

LABEL maintainer="zt-sv"
USER nonroot:nonroot

COPY --from=build --chown=nonroot:nonroot /go/github.com/zt-sv/grafana-annotations-bot/dist/grafana-annotations-bot /app

ENTRYPOINT ["/app"]
