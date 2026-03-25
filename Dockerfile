# builder image
FROM golang:1.25-alpine AS builder
WORKDIR /go/src/scm.swisscom.com/cloud-native/services/keycloak-broker
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o keycloak-broker .

# app image
FROM alpine:3.23
LABEL author="Fabio Berchtold <fabio.berchtold@swisscom.com>"

RUN apk --no-cache add ca-certificates

ENV PATH=$PATH:/app
WORKDIR /app
COPY --from=builder /go/src/scm.swisscom.com/cloud-native/services/keycloak-broker ./keycloak-broker

EXPOSE 8080
CMD ["./keycloak-broker"]
