# Contributing

## 本地开发

```bash
go mod tidy
just test
just build
```

## 运行

```bash
./bin/nacos-cli --help
```

## 提交流程

1. 从 `main` 拉出分支
2. 完成代码与文档修改
3. 本地执行：

```bash
just test
just build
```

4. 提交 PR，等待 CI 通过后合并

## 变更要求

- 保持实现简洁
- 命令示例可直接运行
- 配置字段与实际代码行为保持一致
- 如果行为变化，更新 `README.md` 与 `CHANGELOG.md`
