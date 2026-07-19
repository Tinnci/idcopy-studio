# ID Copy Studio

[![CI](https://github.com/Tinnci/idcopy-studio/actions/workflows/ci.yml/badge.svg)](https://github.com/Tinnci/idcopy-studio/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

本地优先的 Wails 3 + Go 证件复印 PDF 生成器。它把证件正反面图像按真实毫米尺寸排到 A4、Letter、Legal、照片纸或自定义纸张上，并输出不依赖浏览器打印缩放的 PDF。

> 当前使用 Wails `v3.0.0-alpha2.117`。Wails 3 仍是预发布版本，接口和构建模板可能继续变化。

## 已实现

- ISO/IEC 7810 ID-1、ID-2、ID-3、ID-000。
- ICAO Doc 9303 TD1、TD2、TD3。
- 中国大陆、港澳台、日本、韩国、东南亚、印度、欧盟、北美等常见证件预设。
- A3/A4/A5/A6、ISO/JIS B 系列、Letter、Legal、Tabloid、16K、Hagaki、4×6、5×7 和自定义纸张。
- 正反横排、正反纵排、独立网格、每面单独一页。
- 自动选择纸张方向、多份分页、25%–400% 缩放、页边距与间距。
- JPEG/PNG、本地旋转、裁切填满/完整留白/拉伸、可选裁切标记。
- 原生“另存为”对话框和浏览器下载副本两种导出路径。
- 150/300/600 DPI 或保留原图，JPEG/PNG 编码选择。
- 可配置文字、颜色、角度、透明度、字号、居中/平铺水印。
- 水印在导出前压平到证件图像像素，不作为独立 PDF 文本层。
- PDF 标题、作者、主题和关键词等 Unicode 元数据。
- Go 标准库实现的 PDF 1.4 生成器；图像不会上传网络。
- 可复用 Go 包 `github.com/Tinnci/idcopy-studio/idcopy`，UI 只是其 Wails 前端。

## 开发环境

```text
Go        1.26.4
Wails CLI 3.0.0-alpha2.117
Bun       1.3.14
Vite      8.1.5
tsgo      7.0.0-dev.20260707.2
```

项目使用 Bun，不需要 npm；类型检查使用 `tsgo`，不调用 `tsc`。

## 运行与构建

```bash
git clone https://github.com/Tinnci/idcopy-studio.git
cd idcopy-studio

cd frontend
bun install
bun run typecheck
cd ..

go test ./...
wails3 dev
wails3 build
```

构建后的当前平台二进制和 macOS 应用位于：

```text
bin/idcopy-studio
bin/idcopy-studio.app
```

## Go 包用法

```go
request := idcopy.LayoutRequest{
    PaperID:       "iso-a4",
    DocumentID:    "iso-id-1",
    Orientation:   "auto",
    Arrangement:   "paired",
    Sides:         "both",
    Copies:        1,
    ScalePercent:  100,
    MarginMM:      12,
    GapMM:         8,
    Fit:           "cover",
    Front:         idcopy.ImageInput{DataURL: frontDataURL},
    Back:          idcopy.ImageInput{DataURL: backDataURL},
}

pdfBytes, plan, err := idcopy.GeneratePDF(request)
```

尺寸目录可通过 `idcopy.GetCatalog()` 获取，纯排版预览使用 `idcopy.Plan(request)`，不要求先提供图像。

## 是否已有类似项目？

有相邻项目，但未发现成熟项目同时覆盖“国际证件尺寸目录 + 多纸张真实尺寸排版 + 本地桌面 PDF 输出”：

- GitHub 上主流 ID-card 项目大多聚焦 OCR、相机边缘检测或字段识别。
- `Card2PDF` 一类工具主要面向 TCG/游戏卡牌，不包含 ICAO/ISO 证件模型和隐私边界。
- `gofpdf`、Maroto 等是通用 PDF 库，需要应用自己维护证件标准、排版策略和 UI。
- 部分“身份证生成器”用于合成证件图像，与合法复印排版目标不同，也不适合作为本项目基础。

因此本项目的差异点是：不识别或保存身份字段，只做本地像素到物理纸张的可验证排版。

## 隐私与限制

- 前端把图像通过本机 Wails 桥传给 Go；代码不包含上传或遥测。
- 单张图像上限 48 MiB、60 MP，只接受 JPEG/PNG。
- 预设尺寸用于排版，不构成证件标准的法律认证；旧版、临时证件和地方发行版本应实测后使用自定义尺寸。
- 当前不做自动边缘检测、透视校正、OCR、数字签名或打印机 ICC 校色。
- 压平水印能阻止直接删除 PDF 文本层，但不能防止高级图像修复、裁切或重新拍摄。
- 当前输出格式是 PDF；PNG/JPEG 是 PDF 内部图像编码选项，不是独立页面文件导出。

## CI/CD

- `CI`：Linux 上执行 Go 核心测试/静态检查和 Bun/tsgo 前端构建；macOS 上执行完整 Wails 构建。
- `Release`：推送 `v*` tag 后生成 ad-hoc 签名的 macOS `.app` 压缩包、SHA-256 文件和 GitHub Release。
- Dependabot 每周检查 Go、前端和 GitHub Actions 依赖。

公开发布前仍应配置 Apple Developer ID 和 notarization；当前 release workflow 只做 ad-hoc 签名。

更多尺寸来源与采用原则见 [docs/standards.md](docs/standards.md)。
