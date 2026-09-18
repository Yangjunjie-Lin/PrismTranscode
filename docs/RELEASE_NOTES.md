## 流光转码 2.1 — 在线工作台与桌面可靠性升级

- 在线版：浏览器内 FFmpeg WebAssembly，14 个输出预设，批量队列、独立格式、裁剪/质量/尺寸、取消/重试、下载、SHA-256 和 JSON 记录。媒体不上传服务器。
- 桌面版：保留 38 个输出预设和 NCM，新增数据目录独占锁、严格输入校验和失败导入清理。
- Go 1.27.1 构建；跨平台测试、Linux race/vet/漏洞扫描、真实浏览器测试、Windows 原生媒体回归为发布门槛。

下载 `PrismTranscode-Windows-x64-*.zip`，核对 `SHA256SUMS.txt` 后完整解压运行。FFmpeg/FFprobe 不随桌面 ZIP 捆绑，需主动安装或指定已有引擎。`build-manifest.json` 记录构建提交和程序哈希。

在线使用：https://prismtranscode.vercel.app

`FFmpeg-WASM-source-bundle.tar` 提供在线引擎源码、构建配方、链接依赖源码及许可证。自有代码 MIT，WebAssembly 引擎 GPL-2.0-or-later；不是全部 MIT。

**社区预发布 / Beta，Windows EXE 未签名。** 未完成真实 GPU/驱动全矩阵、原生选择器人工验收、全浏览器/移动端、大文件压力测试、独立安全审计和商业许可/专利审查。详见 docs/RELEASE.md。
