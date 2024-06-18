<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a name="readme-top"></a>

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/redds-be/reddlinks">
    <img src="static/assets/img/reddlinks_logo_d.png" alt="Logo" width="128" height="128">
  </a>

<h3 align="center">reddlinks</h3>

  <p align="center">
    A simple link shortener written in Go
    <br />
    <br />
    <a href="https://ls.redds.be">View Demo</a>
    ·
    <a href="https://github.com/redds-be/reddlinks/wiki">Wiki</a>
    ·
    <a href="https://github.com/redds-be/reddlinks/issues">Report Bug</a>
    ·
    <a href="https://github.com/redds-be/reddlinks/issues">Request Feature</a>
  </p>
</div>

<!-- PROJECT SHIELDS -->
![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/redds-be/reddlinks/golangci-lint.yml?label=Golangci-lint)
![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/redds-be/reddlinks/gotest.yml?label=Go%20test)
![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/redds-be/reddlinks/gobuild.yml?label=Go%20build)
![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/redds-be/reddlinks/docker-build.yml?label=Docker%20build)
![GitHub pull requests](https://img.shields.io/github/issues-pr/redds-be/reddlinks)
![GitHub issues](https://img.shields.io/github/issues/redds-be/reddlinks)
![GitHub License](https://img.shields.io/github/license/redds-be/reddlinks)
![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/redds-be/reddlinks)
[![Go Reference](https://pkg.go.dev/badge/github.com/redds-be/reddlinks.svg)](https://pkg.go.dev/github.com/redds-be/reddlinks)

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li><a href="#about-reddlinks">About The Project</a></li>
    <li><a href="#features">Features</a></li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#installation">Installation</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About reddlinks

A simple link shortener written in Go. Made while I was bored.

In case it has been a while since the last commit, no, this project is not dead, if you have an issue or have a feature request, I will respond to it in a reasonable time frame.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Features

- URL shortening
- Front-facing website
- API endpoints
- Random path generation (ex: ls.redds.be/**ag4vb~**, defaults to a pre-configured value)
- Custom path (ex: ls.redds.be/**custom**, overrides path generation)
- Password protected links using argon2
- PostgreSQL and SQLite

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

To use it in a web browser, follow the instruction on the front page.

API requests:

1. Shorten link:

Use your favorite http client to make a POST request in JSON, example with `curl`:

```console
curl -X POST https://ls.redds.be -H 'Content-Type: application/json' -d '{"url":"http://example.com"}'
```
Available params for link shortening are:

- "url": "URL". A valid URL
- "length": "Number". A number for an auto-generated short path, defaults to a pre-configure length. **Optional**
- "customPath": "Path". A custom path to access the shortened link instead of an auto-generated one **Optional**
- "expireAfter": "1d1h1m1s". 1d = 1 day; 1h = 1 hour; 1m = 1 minute; 1s = 1 second; the format should be entered from greater (1d) to lesser (1s). Defaults to a pre-configured time. Example : "3d5h34m54s" = 3 days, 5 hours, 34 minutes and 54 seconds from now. **Optional**
- "password": "Password". A password to protect the shortened link with. **Optional**

2. Access password-protected links:

Use your favorite http client to make a GET request whilst posting JSON, example with `curl`:

```console
curl -X GET https://ls.redds.be/ag4vb~ -H 'Content-Type: application/json' -d '{"password":"secret123"}'
```

More information in the [wiki](https://github.com/redds-be/reddlinks/wiki/Usage).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Installation

Read the configuration and installation instructions in the [wiki](https://github.com/redds-be/reddlinks/wiki/Installation).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ROADMAP -->
## Roadmap

- [ ] Add a private mode with authentication

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->
## Contributing

I don't expect anyone other than me to contribute, but you should follow these steps :

**Fork -> Patch -> Push -> Pull Request**

The **Go** code is linted with [`golangci-lint`](https://golangci-lint.run) and
formatted with [`golines`](https://github.com/segmentio/golines) (width 120) and
[`gofumpt`](https://github.com/mvdan/gofumpt). See the Makefile targets.
If there are false positives, feel free to use the
[`//nolint:`](https://golangci-lint.run/usage/false-positives/#nolint-directive) directive
and justify it when committing to your branch or in your pull request.

For any contribution to the code, make sure to create tests/alter the already existing ones according to the new code.

Make sure to run `make prep` before committing any code.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

*Project under the [GPLv3 License](https://www.gnu.org/licenses/gpl-3.0.html).*

*Copyright (C) 2024 redd*

<p align="right">(<a href="#readme-top">back to top</a>)</p>
