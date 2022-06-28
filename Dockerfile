FROM golang:1.18 as dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY ./app ./app

RUN CGO_ENABLED=0 go build -o /builded_app_a ./app/cmd/app_a/main.go \
&& CGO_ENABLED=0 go build -o /builded_app_b ./app/cmd/app_b/main.go


FROM alpine:latest as app_a
COPY --from=dev /builded_app_a /builded_app_a
EXPOSE 8091
CMD ["/builded_app_a"]


FROM alpine:latest as app_b
COPY --from=dev /builded_app_b /builded_app_b
EXPOSE 8092
CMD ["/builded_app_b"]
