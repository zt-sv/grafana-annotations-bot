ARG go_version=1.21.6
ARG app_version=dev
FROM golang:${go_version}-alpine AS build

ENV GO111MODULE=auto
ENV GOBIN=/go/bin
ENV APP_VERSION=$app_version

RUN apk --no-cache add make

WORKDIR /go/github.com/zt-sv/grafana-annotations-bot
ADD . ./
RUN make build

FROM gcr.io/distroless/static AS final

LABEL maintainer="zt-sv"
USER nonroot:nonroot

COPY --from=build --chown=nonroot:nonroot /go/github.com/zt-sv/grafana-annotations-bot/dist/grafana-annotations-bot /app

ENTRYPOINT ["/app"]
