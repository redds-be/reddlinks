# Deploy reddlinks with docker

## Docker compose

`docker compose` can be used to automate the building and running of reddlinks.

By default, the docker-compose file uses the port 8080 (- "8080:8080"), feel free to change it for a more convenient one.
/!\ Only change the first part of the ports statement (ex: "8081:8080").

To build and run reddlinks using `docker compose`, change directory to `docker/` and run:

```console
docker compose build --build-arg tag_version=$(git describe --tags)
docker compose up
```

> Note: You could only use `docker compose up` if you don't want to display the version information.

You can also use the `-d` option to run reddlinks as a daemon:

```console
docker compose build --build-arg tag_version=$(git describe --tags)
docker compose up -d
```

> Note: You could only use `docker compose up -d` if you don't want to display the version information.

In case of errors, you can use `docker compose logs -f` to see the logs (it is recommended to use it when you start reddlinks for the first time to see if everything works):

```console
docker compose build --build-arg tag_version=$(git describe --tags)
docker compose up -d && docker compose logs -f
```

> Note: You could only use `docker compose up -d && docker compose logs -f` if you don't want to display the version information.
