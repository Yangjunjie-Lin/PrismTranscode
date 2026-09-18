# 流光转码 · PrismTranscode 桌面使用说明

本文件主体保留 2.0 功能说明；2.1 新增的在线版、实例锁与本次 Windows 测试结果请见 [README](README.md) 和 [发布检查](docs/RELEASE.md)。本机服务不可作为 Vercel 后端部署。

一个离线优先、本地运行的批量媒体转换工作台。由原 NCM→MP3 工具升级，支持音频、视频、静态图片、GIF 与文本字幕。Windows 主程序使用 Go 编译，启动后打开本机浏览器；**不是云端转换网站，也不是原生 WebView 窗口**。

**交付状态：可运行工程版。** 真实转换核心已在 Linux / FFmpeg 7.1.5 上回归；Windows x64 EXE 已交叉编译，尚未进行 Windows 实机验收。这里的 38 是“输出预设数”，不是声称支持任意文件、任意编码组合。详细结果与未验证项见 `docs/ACCEPTANCE.md` 和交付包的 `测试报告/`。

## 1. Windows 开始使用

完整解压便携包，双击 `PrismTranscode.exe`。不需要安装 Python、Go 或 .NET；需要系统浏览器，以及可运行的 FFmpeg 和 FFprobe 引擎。

首次打开后进入“转换引擎”。已有引擎时，选择同时包含 `ffmpeg.exe`、`ffprobe.exe` 的文件夹；或者点击“安装完整引擎 · BtbN”。程序只在明确确认安装时下载外部组件，按发行方 SHA-256 校验后安装。下载失败不会变成“已安装”。主程序包**不附带 FFmpeg 二进制**，因此首次获取引擎需要联网；也可以在其他设备下载并复制两个程序及其发行说明到主程序旁 `tools/`。

BtbN 安装项使用完整的 GPL 开发构建；Gyan 安装项使用 release essentials，两者不能理解为同一版本或同一能力。Essentials 通常缺少 SoXR、SVT-AV1 等组件，相应预设会禁用。追求可重复生产环境时，手动选择经自己验收的**固定版本完整引擎**，并保存其哈希与构建配置，不要每次追踪 latest。Windows 建议使用 Windows 10/11 x64；实际系统要求仍由所选引擎决定。

通过“添加文件”“添加文件夹”或“路径导入”读取本地文件。Windows 原生选择器直接读取路径；拖放/浏览器上传会把文件复制到本机导入缓存，不会发送到互联网。大型视频优先使用直接路径，避免双份磁盘占用。

导入前设置右侧默认目标；导入后，每行可以独立改输出格式。右侧高级参数只在点击“将以上参数应用到选中文件”时修改已有任务。填写输出目录，选择任务，点击“开始转换”。不同格式可以同批执行，例如视频→MP3、WAV→FLAC、PNG→WebP、SRT→VTT。

完成后打开输出目录；详情里有实际识别信息、最终执行参数数组、警告、输出 SHA-256 和引擎日志。失败文件不会终止其他任务。“选择失败 / 停止项”后可以重试；已经完成的文件要重新转换，先重新应用参数。

关闭网页不会结束后台程序，使用右上角“退出”。默认设置和队列保存在 `%APPDATA%\PrismTranscode`。浏览器未打开时，可查看该目录 `last-session-url.txt` 中的本次本机地址；地址仅在对应进程运行期间有效。不要分享其中的会话令牌。

## 2. 已实现的能力

| 类别 | 已实现 |
| --- | --- |
| 输入识别 | 根据实际容器与媒体流识别，不仅查看扩展名；显示编码、时长、尺寸、采样率、声道、位深与 HDR 标志。NCM 先识别签名并解码，再探测内部媒体。 |
| 批处理 | 混合格式导入、去重、文件夹递归扫描、逐文件目标格式、选中任务统一参数、1–4 个并行任务、进度、停止和失败重试。 |
| 音频输出 | MP3、M4A/AAC、AAC/ADTS、FLAC、WAV/PCM、M4A/ALAC、AIFF、OGG/Vorbis、Opus、WMA、AC-3、WavPack、原码音轨提取。 |
| 视频输出 | H.264、H.265/HEVC、AV1、VP9、ProRes 422 HQ、FFV1；对应 MP4、MKV、WebM、MOV、AVI、MPG、TS 预设；MP4/MKV 原码重封装。 |
| 图片与字幕 | PNG、JPEG、WebP、无损 WebP、AVIF、无损 JPEG XL、TIFF、BMP、GIF；SRT、WebVTT、ASS。静态图片预设只取一帧。 |
| 精度与参数 | 优先原码音频提取；保留采样参数；按需求 SoXR 重采样；位深/声道/码率设置；视频质量、缩小上限、帧率、起点和时长；GPU 编码及记录在案的 CPU 回退。 |
| 安全交付 | 原文件只读、不覆盖已有输出、临时文件完成后提交、输出重新探测、默认完整解码检查、可选 SHA-256、JSON/CSV 记录、队列恢复。 |

更细的编码/容器及保真边界见 [格式与质量矩阵](docs/FORMATS_QUALITY.md)。

## 3. 三个常见操作

**视频转 MP3：** 导入 MP4/MKV/MOV 等具有音轨的视频 → 该行选 MP3 → 输出目录 → 开始。默认高质量 MP3 转码采用 320 kbps；若输入本来就是 MP3 音轨且无需变换，则直接复制压缩音频。没有音轨的视频会报错，不生成假的 MP3。

**尽可能无损提取视频声音：** 选“原码提取 · 不重编码”。例如 AAC 音轨通常输出 M4A，MP3 音轨输出 MP3。原码提取不承诺强制得到 MP3；AAC 变成 MP3 必须重新编码。

**批量 NCM 转不同格式：** 导入 NCM，选择 MP3/FLAC/WAV 等。内部已是 MP3 时默认尽可能直接提取；MP3→FLAC 只改变存储方式，不恢复已经损失的信息。NCM 封面本版不迁移，标题、歌手、专辑文字可保留。

## 4. 运行与构建源码

最低语言基线是 Go 1.23；正式发布应使用当前受支持的 Go 稳定版。项目没有第三方 Go 模块，不需要 npm 安装前端依赖。

```sh
go test ./...
go run . --no-browser --ffmpeg /usr/bin/ffmpeg --data-dir ./dev-data
# 终端显示含会话令牌的 http://127.0.0.1:随机端口/app/... 地址
```

Windows 编译：运行 `scripts/build_windows.ps1`，或：

```powershell
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -trimpath -buildvcs=false -ldflags "-s -w -H=windowsgui" -o PrismTranscode.exe .
```

Linux/macOS 从源码可构建主程序，文件输入使用路径或浏览器本地导入；内置自动下载器只支持 Windows x64。Linux/macOS 需要自行提供本机 FFmpeg/FFprobe。跨平台编译不等于所有平台已经实机测试。

`scripts/smoke_test.py` 生成合成媒体并驱动真实 HTTP API 与 FFmpeg；安装 Python、Go、FFmpeg 后可复现。`ui_test.py` 用于标准浏览器端到端测试；受限浏览器环境可用 `ui_memory_test.py` 做内存页面组件测试，但不能把后者当作原生浏览器导航验收。详见 [验收说明](docs/ACCEPTANCE.md)。

## 5. 工程结构

```text
main.go                   本地 HTTP 应用、鉴权、导入、设置、导出
internal/ncm/             NCM 容器解析与流式解码
internal/media/           FFprobe 探测、编码能力、参数规划、执行与校验
internal/jobs/            持久化任务队列、并发、停止、失败隔离
internal/installer/       显式引擎安装、下载白名单、SHA-256、安装收据
internal/platform/        Windows 原生文件选择与系统打开功能
web/                      无前端构建依赖的 HTML/CSS/JavaScript 界面
cmd/fixtures/             生成有合法测试内容的 NCM 固件
scripts/                  构建、核心集成测试、界面测试
.github/workflows/        跨平台单元测试、Windows 构建、依赖扫描配置
```

[架构与扩展](docs/ARCHITECTURE.md) · [API](docs/API.md) · [安全与隐私](SECURITY.md) · [开源来源](THIRD_PARTY_NOTICES.md) · [发布检查](docs/ACCEPTANCE.md)

## 6. 明确不包含

不包含 DOCX/PDF/电子表格互转、压缩包转换、RAW 专业显影、PGS 图像字幕 OCR、云端视频下载、Widevine/FairPlay 解锁或 QMC/KGM 等其他受保护平台格式解密。HEIC/HEIF、老旧或特殊输入能否读取取决于所选引擎与具体文件；它们不是已经全面验证的输入承诺。

没有 AI 补帧、AI 超分、音频“无损修复”等生成式处理。普通视频转码不保证全部字幕/封面/附件与私有元数据完整迁移；保留完整轨道优先选择兼容的原码重封装。HDR 默认阻止普通有损转码，不静默输出错误颜色。所有无损表述均有明确范围，不能理解为容器每个字节或全套元数据保持一致。

这是本机单用户工具，不是公开网络服务。请勿同时用多个进程写同一个数据目录，也不要对外开放或反向代理其接口。

## 许可

本项目自有代码采用 MIT 许可，保留原 NCM-MP3 Batch 贡献者声明。Go 运行时及外部 FFmpeg/编码器各遵守自身许可。本包未合并 LosslessCut 或 Shutter Encoder 的 GUI 源码，也未把 GPL 组件改称 MIT；引用它们的工作流设计与实际调用的外部引擎是两回事。商用分发前需独立审查所附二进制、许可证、专利与媒体授权要求。
