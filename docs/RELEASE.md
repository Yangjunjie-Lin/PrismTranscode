# 2.1.0-beta.1 发布范围与复现

## 发布定位

本版本为社区预发布，可实际使用，但不标记为已认证商业稳定版。2.0 历史验收见 ACCEPTANCE.md；不要把历史 Linux 回归当成本次 Windows 证据。

本次 Windows x64 / Go 1.27.1 / BtbN FFmpeg N-126626-g7070fe638e-20260917 实测原生集成 **59/59** 通过，覆盖全部 38 个输出预设、PCM/像素/原码包一致性、取消、失败隔离、无覆盖、队列恢复及无 GPU 时的 CPU 回退。未使用真实用户歌曲。本次 FFmpeg 下载归档 SHA-256：`fa7dd8aef59d47414f511e209f2964c0327e83dfa147753ef65cc98b98c7c04f`。

Go 单元测试、vet、govulncheck 和 npm audit 已本地运行。浏览器测试使用实际 HTTP 页面及真实 WebAssembly，不是内存页面模拟，覆盖转换、下载哈希、损坏文件隔离、停止重试、NCM 拒绝、文件名注入和窄屏布局。CI 是后续变更的持续验证来源。

在线版 7 项单元测试、5 组浏览器 E2E（含 14/14 预设真实转换）通过；桌面 Windows 浏览器 GUI 验收 12/12 通过。脱敏证据见 release-evidence.json。

2026-09-18：已部署 https://prismtranscode.vercel.app，并在生产 HTTPS/CSP 环境重新运行全部 5 组浏览器 E2E，通过。源码仓库已连接 Vercel，推送 main 会自动触发部署。

## 本地验证

```sh
go test ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
npm ci
npm run check
npm audit --audit-level=high
npx playwright install chromium
npm run test:e2e
```

Windows 真实引擎测试：`scripts/install-test-engine.ps1` 按发行方 SHA-256 下载到项目 `.cache`，不会全局安装。将其 bin 和 Go 加入当前进程 PATH 后，构建控制台版并运行 `python scripts/smoke_test.py --binary artifacts/PrismTranscode-console.exe --work .cache/new-smoke --report artifacts/windows-smoke-report.json`。工作目录每次应不同。

## 发布步骤

1. 确认 main CI 全绿，检查依赖/许可变更，无令牌、用户媒体、运行数据和本机路径日志被提交。
2. 更新主程序、npm package、界面、安装器 User-Agent、CHANGELOG、文档版本。
3. 创建 `v2.1.0-beta.1` 类标签。Release 工作流复用完整 CI，再打包 Windows ZIP；GitHub 自动提供该标签的源码档案。
4. 运行 `node scripts/bundle-wasm-sources.mjs`，使用 `tar -cf artifacts/FFmpeg-WASM-source-bundle.tar -C artifacts/wasm-sources .` 打包。上传到同一 Release，提供 GPL 核心、构建配方和依赖源码来源，保留许可证和摘要；其配方含上游分支引用，不声称字节级重现 npm 二进制。
5. 使用 `vercel --prod` 部署 `source` 仓库根目录。项目配置将 `online/dist` 作为输出，不部署 Go 后端。无需环境变量、数据库或付费媒体处理服务。Vercel 用量仍受账户计划限制。
6. 对实际生产 HTTPS 域名执行 `TEST_BASE_URL=https://... npm run test:e2e`；确认网站、引擎、下载、CSP 和许可文件能公开访问。

## 仍未完成的正式商业发行条件

在线版固定核心的 libopus/VP9 编码在本次 Chromium 回归中出现 WASM 越界，因此不使用它们：Opus 预设明确标记为 FFmpeg 实验性编码器、固定 48 kHz 立体声，WebM 使用 VP8 + Vorbis。桌面版仍按原生引擎能力提供 libopus/VP9。在线探测、编码、输出验证阶段隔离 Worker 内存，避免失败传播。

- Windows Authenticode 代码签名证书与签名发行；当前 EXE **未签名**。
- NVIDIA/Intel/AMD 真实硬件和驱动矩阵、原生选择器人工验收、SmartScreen 行为、多版本 Windows 与大文件/磁盘满/断电压力测试。
- 原生媒体进程隔离、安全审计，以及 WebAssembly 内部 C/C++ 依赖的漏洞处置；npm audit 不覆盖这些库。
- 独立 GPL、专利和商用分发法律审查。源代码档案及许可通知不等同于法律意见或二进制可复现认证。
- 移动浏览器、Firefox/Safari 的完整实测；首发端到端基线为桌面 Chromium。

本次发布不声称以上条件已经满足。
