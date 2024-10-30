# GoGrip - GitHub Readme Preview

A minimal command-line tool that renders local markdown files using GitHub's API.

## Installation

You can install via go using the following command:

```shell
go install github.com/sixlive/gogrip@latest
```

You can also install by snagging a pre-built binary from the [releases](https://github.com/sixlive/gogrip/releases/latest) page.

## Usage

Basic usage with default README.md:
```shell
gogrip
```

Specify a different file:
```shell
gogrip -f CONTRIBUTING.md
```

Open in browser automatically:
```shell
gogrip -b
```

Custom host and port:
```shell
gogrip -host 0.0.0.0 -port 8080
```

```shell
With GitHub authentication (for higher rate limits):
gogrip -token YOUR_GITHUB_TOKEN
```

## Options

```
Usage of gogrip:
  -host string
        Host to listen on (default "localhost")
  -port int
        Port to listen on (default 6419)
  -f string
        File to render (default "README.md")
  -token string
        GitHub personal access token
  -b    
        Open browser automatically
```

## Acknowledgments

This project is inspired by [Grip](https://github.com/joeyespo/grip), the Python-based GitHub Readme Instant Preview tool.

## License

MIT
