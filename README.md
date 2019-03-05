# byebye-slack-channel

A simple bot that notifies when a specific manga is released on MangaStream.

 * [Installation](README.md#installation)
      * [Binaries](README.md#binaries)
      * [Via Go](README.md#via-go)
      * [Running with Docker](README.md#running-with-docker)
 * [Usage](README.md#usage)

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/isaacasensio/byebye-slack-channel/releases).

#### Via Go

```console
$ go get github.com/isaacasensio/byebye-slack-channel
```

#### Running with Docker

```console
docker run -d \
    -v $(pwd)/token:/token \
    -v $(pwd)/exclude.conf:/exclude.conf \
    --name byebye-slack-channel \
    isaacasensio/byebye-slack-channel:0.0.1 \
    --dry-run=true \
    --exclude-channels-path=/exclude.conf \
    --token-path=/token \
    --user-id=UC7HFR88V
```

## Usage

```console
$ byebye-slack-channel -h
byebye-slack-channel -  A simple tool to leave slack channels in bulk.

Usage: byebye-slack-channel <command>

Flags:

  --token-path                 path to a file that contains your Slack token
  --exclude-channels-path      path to a file that contains a list of channels to exclude
  --user-id                    slack userID
  --dry-run                    runs command without removing the channels

Commands:

  version  Show the version information.
```
