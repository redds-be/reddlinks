FROM golang:1.22-alpine3.19 as build

RUN apk update && apk add git

WORKDIR /go/src/reddlinks
COPY . .

RUN go mod download 
RUN version=$(git describe --tags) && sed "s/noVersion/$version/g" main.go > main.go.alt && mv main.go.alt main.go
RUN CGO_ENABLED=0 go build -o /go/bin/reddlinks

FROM gcr.io/distroless/base-debian12

COPY --from=build /go/bin/reddlinks /reddlinks

EXPOSE 8080

ENTRYPOINT ["/reddlinks"]
