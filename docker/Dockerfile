FROM golang:1.21-alpine3.19

RUN apk update && apk add --no-cache build-base git

EXPOSE 8080

RUN mkdir /opt/reddlinks
RUN adduser -D rluser

WORKDIR /opt/reddlinks

COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod vendor
COPY . .

RUN chown -R rluser:rluser /opt/reddlinks
RUN git config --global --add safe.directory /opt/reddlinks
RUN make
RUN mv build/reddlinks /usr/local/bin/

ENTRYPOINT ["reddlinks"]