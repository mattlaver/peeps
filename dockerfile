# Build the Go Binary.

FROM golang:1.12.1 as build
ENV CGO_ENABLED 0
ARG VCS_REF
ARG PACKAGE_NAME
ARG PACKAGE_PREFIX
RUN mkdir -p /go/src/github.com/mattlaver/peeps
COPY . /go/src/github.com/mattlaver/peeps
WORKDIR /go/src/github.com/mattlaver/peeps/cmd/${PACKAGE_PREFIX}${PACKAGE_NAME}
RUN go build -ldflags "-s -w -X main.build=${VCS_REF}" -a -tags netgo


# Run the Go Binary in Alpine.

FROM alpine:3.7
ARG BUILD_DATE
ARG VCS_REF
ARG PACKAGE_NAME
ARG PACKAGE_PREFIX
COPY --from=build /go/src/github.com/mattlaver/peeps/cmd/${PACKAGE_PREFIX}${PACKAGE_NAME}/${PACKAGE_NAME} /app/main
COPY --from=build /go/src/github.com/mattlaver/peeps/private.pem /app/private.pem
WORKDIR /app
CMD /app/main

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="${PACKAGE_NAME}" \
      org.opencontainers.image.authors="Matt Laver <matt@laver.io>" \
      org.opencontainers.image.source="https://github.com/mattlaver/peeps/cmd/${PACKAGE_PREFIX}${PACKAGE_NAME}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.vendor="Laver IO"