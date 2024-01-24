# Deploy reddlinks with docker

It is strongly recommended to use [docker compose]()

## Docker

The dockerfile in this directory can be used to build a docker container for reddlinks.

By default, the Dockerfile uses the port 8080 (EXPOSE 8080), feel free to change it for a more convenient one.

From reddlinks's root directory:

```console
docker build -t reddlinks -f docker/Dockerfile .
```

From the `docker/` directory:

```console
docker build -t reddlinks -f Dockerfile ../
```

Run the container:

```console
docker run reddlinks --user 1000:1000
```

## Docker compose

`docker compose` can be used to automate the build and run step reddlinks.

By default, the docker-compose file uses the port 8080 (- "8080:8080"), feel free to change it for a more convenient one.
/!\ Only change the first part of the ports statement (ex: "8081:8080").

To build and run reddlinks using `docker compose`, change directory to `docker/` and run:

```console
docker compose up
```

You can also use the `-d` option to run reddlinks as a daemon:

```console
docker compose up -d
```

In case of errors, you can use `docker compose logs -f` to see the logs (it is recommended to use it when you start reddlinks manually):

```console
docker compose up -d && docker compose logs -f
```