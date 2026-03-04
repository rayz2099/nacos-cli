# nacos-cli

English | [中文](./README-zh.md)

nacos-cli is a command-line tool for operating Nacos, covering common configuration management and service discovery workflows.

## Install

### 1) Install with `go install`

```bash
go install .
```

After installation:

```bash
nacos-cli --help
```

### 2) Build locally (via justfile)

```bash
just build
```

Binary output:

```bash
./bin/nacos-cli
```

## Quick Start

The examples below assume your Nacos server is reachable and required config is set.

```bash
nacos-cli config list
nacos-cli config get dt-rpc COMMON
nacos-cli config get dt-rpc
nacos-cli naming instances --service demo-service
```

Notes:
- `group` defaults to `COMMON` in `config get [data-id] [group]`
- `naming instances` requires `--service`

## Configuration

Config file path:

```text
~/.config/nacos-cli/config.json
```

Supported fields:

- `nacos_server_addr`
- `nacos_username`
- `nacos_password`
- `nacos_namespace`
- `namespaces`
- `nacos_output`

Example:

```json
{
  "nacos_server_addr": "127.0.0.1:8848",
  "nacos_username": "",
  "nacos_password": "",
  "nacos_namespace": "public",
  "namespaces": ["public"],
  "nacos_output": "text"
}
```

Priority:

```text
flags > env > file > default
```

Environment variables (both lowercase and uppercase are supported):

- `nacos_server_addr` / `NACOS_SERVER_ADDR`
- `nacos_username` / `NACOS_USERNAME`
- `nacos_password` / `NACOS_PASSWORD`
- `nacos_namespace` / `NACOS_NAMESPACE`
- `nacos_output` / `NACOS_OUTPUT`

## Output Modes

- `text` (default)
- `json`

Set via global flag:

```bash
nacos-cli -o json config list
```

## Runtime Cache and Log

- Cache directory: `~/.config/nacos-cli/cache`
- Log is disabled by default
- Enable dev mode with global flag `--dev`
- In dev mode, log directory: `~/.config/nacos-cli/log` with `debug` level

Example:

```bash
nacos-cli --dev config list
```

## Fish Completion

Enable fish completion:

```bash
mkdir -p ~/.config/fish/completions
nacos-cli completion fish > ~/.config/fish/completions/nacos-cli.fish
```

