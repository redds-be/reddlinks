FROM golang:1.22-alpine3.19 as build

ARG tag_version=noVersion
ENV tag_version=$tag_version

WORKDIR /go/src/reddlinks
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-X 'main.Version=$tag_version'" -o /go/bin/reddlinks

FROM gcr.io/distroless/base-debian12

COPY --from=build /go/bin/reddlinks /reddlinks

EXPOSE 8080

ENTRYPOINT ["/reddlinks"]
