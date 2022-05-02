FROM golang:1.17-alpine as build
WORKDIR "/build"
COPY .  /build
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /build/kube-metric-sender -a -ldflags '-extldflags "-static"' .

FROM debian:bullseye-slim AS certs

RUN \
  apt update && \
  apt install -y ca-certificates && \
  cat /etc/ssl/certs/* > /ca-certificates.crt


FROM  scratch as final

COPY --from=build  /build/kube-metric-sender  /kube-metric-sender
COPY --from=certs /ca-certificates.crt /etc/ssl/certs/
WORKDIR "/"
ENTRYPOINT ["/kube-metric-sender"]
