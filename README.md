# terrain-edit

地块编辑：权限校验、编辑幂等、撤销栈、按地块索引，只用 Go 标准库。

```
go test ./...
go run ./cmd/terrain-edit --sample permission
go run ./cmd/terrain-edit --sample duplicate
go run ./cmd/terrain-edit --sample undo
```

## 口径（README 为准）

- **权限**：只有角色是 `admin` 或 `builder` 的玩家能编辑，其他人返回 `(false, "denied")`
  且不改状态。
- **编辑幂等**：把某个地块设成它当前的值属于重复编辑，返回 `(false, "duplicate")` 且不入撤销栈。
- **可撤销**：`Undo` 按后进先出恢复上一次编辑前的值并返回 `true`，栈空返回 `false`。
- **查询代价**：`Value` 不许每次全扫地块表，`Scanned` 不随地块数线性增长。
- **规模**：单张地图 200 万格、每秒 10 万次编辑与查询，单次摊还 O(1)，内存 O(地块数)。

## 输出契约（不改格式）

```
ok=false reason=denied
second=false reason=duplicate
ok=true value=0
```

## 目录

```
editor.go               权限、幂等与撤销
cmd/terrain-edit        命令行入口
editor_test.go          go test 用例
```
