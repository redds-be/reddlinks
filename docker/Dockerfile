FROM golang:1.22-alpine3.19 as build

RUN apk update && apk add --no-cache build-base git

WORKDIR /go/src/reddlinks
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 make
RUN mv build/reddlinks /go/bin/reddlinks

FROM gcr.io/distroless/base-debian12

COPY --from=build /go/bin/reddlinks /reddlinks
COPY --from=build /go/src/reddlinks/static /static

EXPOSE 8080

ENTRYPOINT ["/reddlinks"]
