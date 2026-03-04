# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Initial open-source release documentation:
  - `README.md`
  - `LICENSE` (MIT)
  - `CONTRIBUTING.md`
  - GitHub Actions CI workflow

### Release checklist
- [ ] 示例命令可复制执行
- [ ] Completion 可用
- [ ] License 存在
- [ ] CI 通过
- [ ] 首个 tag 命名：`v0.1.0`

## [0.1.0] - 2026-03-04

### Added
- Core CLI commands for Nacos operations:
  - `config` operations (`get/put/delete/list`)
  - `naming` operations (`register/deregister/instances`)
- Runtime config resolution with priority:
  - `flags > env > file > default`
- Output formats:
  - `text`
  - `json`
