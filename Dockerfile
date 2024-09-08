FROM golang:1.23-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o zadanie-6105 cmd/main.go

FROM alpine
WORKDIR /etc/zadanie-6105
ENV PATH=/etc/zadanie-6105:${PATH}
COPY --from=build /src/zadanie-6105 .

ENTRYPOINT ["zadanie-6105"]
