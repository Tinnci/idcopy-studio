# Contributing

感谢参与 ID Copy Studio。

## 开发规则

- 使用 Go 和 Wails 3；前端依赖使用 Bun。
- TypeScript 检查使用 `tsgo`，不要调用 `tsc`。
- 不上传、记录或持久化用户证件图像。
- 新增证件尺寸时，应同时记录标准名称、地区、毫米尺寸和来源依据。
- PDF 或排版变更必须增加回归测试。

## 本地检查

```bash
go test ./...
go vet ./...

cd frontend
bun install
bun run typecheck
bun run build
```

## 提交内容

Pull request 应说明：

- 修改了什么以及原因；
- 是否影响物理尺寸或 PDF 输出；
- 使用了哪些标准或公开来源；
- 完成了哪些验证。
