---
title: "Installation"
description: "Get the csdn binary."
weight: 20
---

## Homebrew

```bash
brew install tamnd/tap/csdn
```

## Pre-built binaries

Download the archive for your platform from
[Releases](https://github.com/tamnd/csdn-cli/releases), extract it, and place
`csdn` on your `$PATH`. Each release carries archives for Linux, macOS, Windows,
and FreeBSD across amd64 and arm64, with a signed `checksums.txt` you can verify
with keyless [cosign](https://docs.sigstore.dev/).

## Go

```bash
go install github.com/tamnd/csdn-cli/cmd/csdn@latest
```

That puts `csdn` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless you moved
it. Make sure that directory is on your `$PATH`.

## Docker

```bash
docker run --rm ghcr.io/tamnd/csdn:latest hot
```

## Linux packages

`deb`, `rpm`, and `apk` packages are available on the
[Releases](https://github.com/tamnd/csdn-cli/releases) page.

## Shell completion

```bash
csdn completion bash    # or zsh, fish, powershell
```

Run `csdn completion <shell> --help` for where to install the script.

## Checking the install

```bash
csdn version
```

prints the version and exits.
