# nacos-cli

[English](./README.md) | 中文

nacos-cli 是一个用于在命令行中操作 Nacos 的工具，支持配置管理与服务发现常用操作。

## 安装

### 1) 使用 `go install`

```bash
go install .
```

安装后可直接使用：

```bash
nacos-cli --help
```

### 2) 本地构建（复用 justfile）

```bash
just build
```

二进制文件输出到：

```bash
./bin/nacos-cli
```

## 快速开始

以下示例默认你已可访问 Nacos，并已完成必要配置。

```bash
nacos-cli config list
nacos-cli config get dt-rpc COMMON
nacos-cli config get dt-rpc
nacos-cli naming instances --service demo-service
```

说明：
- `config get [data-id] [group]` 中 `group` 默认值为 `COMMON`
- `naming instances` 需要通过 `--service` 指定服务名

## 配置

配置文件路径：

```text
~/.config/nacos-cli/config.json
```

支持字段：

- `nacos_server_addr`
- `nacos_username`
- `nacos_password`
- `nacos_namespace`
- `namespaces`
- `nacos_output`

示例：

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

配置优先级：

```text
flags > env > file > default
```

相关环境变量（支持小写与大写）：

- `nacos_server_addr` / `NACOS_SERVER_ADDR`
- `nacos_username` / `NACOS_USERNAME`
- `nacos_password` / `NACOS_PASSWORD`
- `nacos_namespace` / `NACOS_NAMESPACE`
- `nacos_output` / `NACOS_OUTPUT`

## 输出模式

- `text`（默认）
- `json`

可通过全局参数设置：

```bash
nacos-cli -o json config list
```

## Fish Completion

启用 fish completion：

```bash
mkdir -p ~/.config/fish/completions
nacos-cli completion fish > ~/.config/fish/completions/nacos-cli.fish
```

重新打开 shell 后生效。
