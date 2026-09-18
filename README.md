# 流光转码 · PrismTranscode

本地优先的媒体转换工作台：Windows 桌面版 + 无需上传媒体的浏览器在线版。

[在线使用](https://prismtranscode.vercel.app) · [下载 Windows 版](https://github.com/Yangjunjie-Lin/PrismTranscode/releases) · [桌面使用说明](DESKTOP.md) · [安全说明](SECURITY.md)

当前版本：**2.1.0-beta.1，社区预发布**。不是已签名、经独立安全审计或全硬件认证的商业发行版。

## 选择适合你的版本

| | 在线版 | Windows 桌面版 |
|---|---|---|
| 处理方式 | 浏览器内 FFmpeg WebAssembly Worker | 本机 Go + 外部 FFmpeg/FFprobe |
| 输出预设 | 14 个：MP3/WAV/FLAC/M4A/OGG/Opus、MP4/WebM、GIF/PNG/JPEG/WebP、SRT/VTT | 38 个，包含 HEVC、AV1、ProRes、FFV1、AVIF、JXL、重封装等 |
| 文件范围 | 单文件 ≤128 MiB，队列 ≤384 MiB / 50 项，串行处理 | 本地路径、大文件、NCM、1–4 并行任务 |
| 结果 | 逐个下载，JSON 记录、SHA-256、结构检查 | 本机输出目录、JSON/CSV、完整解码检查、SHA-256 |
| 隐私 | 媒体不上传；刷新会丢失队列与结果 | 本地队列持久化；不对外开放后端 |

输出预设不是“任意文件转任意格式”保证。无音轨的视频不能转音频，文本字幕不包含 OCR。在线版不做 HDR 色调映射或 GPU 编码，视频只保留第一条画面/音轨。无损格式不能恢复源文件已丢失的信息。

## 2.1 升级内容

- 新增在线工作台：拖放、多文件去重、逐文件格式、质量设置、裁剪、尺寸上限、进度、取消/重试、下载与 JSON 导出。
- 固定 WebAssembly 引擎，同源部署，无运行时 CDN；检测编码器能力，验证输出结构并计算 SHA-256。
- 桌面数据目录 OS 独占锁，崩溃自动释放；无效上传清理暂存文件；JSON 拒绝尾随数据；启动时不静默替换指定引擎。
- Windows 真实媒体回归、跨平台测试、漏洞扫描、浏览器端到端测试、打包校验清单及 GitHub Release 工作流。

## 本地开发

桌面版（本次验证工具链 Go 1.27.1，外部 FFmpeg/FFprobe 必须另装）：

```sh
go test ./...
go vet ./...
go run . --no-browser --ffmpeg /path/to/ffmpeg --data-dir ./dev-data
```

在线版（Node.js ≥22.12）：

```sh
npm ci
npm run build
npm run dev
npm test
npx playwright install chromium
npm run test:e2e
```

首次启动开发服务器前需运行一次 build，将锁定版本引擎复制到同源静态目录。生产输出为 `online/dist`。

## 测试与发布

```powershell
./scripts/build_windows.ps1
./scripts/package_windows.ps1
```

GitHub Actions 在 Windows/macOS/Linux 执行 Go 测试和 vet、Linux race；另外运行 govulncheck、npm audit、真实浏览器转换及 Windows 原生 FFmpeg 回归。`v*` 标签触发发布流水线，全部检查通过后构建 ZIP、SHA256SUMS 和构建清单；含连字符的版本标记为预发布。

Vercel 只部署静态在线版，配置见 `vercel.json`。不得将桌面本地 API 公开反代为在线服务。详见 [部署与发布指南](docs/RELEASE.md)。

## 隐私与许可

媒体在设备上处理，站点/引擎请求仍受托管商常规访问日志约束。桌面 ZIP 不内置 FFmpeg。在线版引擎基于 FFmpeg 5.1.4，不等于最新版原生 FFmpeg，也未经完整原生库漏洞审计。

自有代码 MIT，保留原 NCM-MP3 Batch 声明。在线版 FFmpeg 核心为 **GPL-2.0-or-later**，其他组件遵循各自许可证；源码包、构建配方和许可全文见 [第三方说明](THIRD_PARTY_NOTICES.md)。不提供独立法律、专利或安全认证。测试素材均为合成内容。
