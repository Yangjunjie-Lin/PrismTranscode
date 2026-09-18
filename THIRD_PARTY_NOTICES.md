# 技术来源与第三方说明

**简体中文** · [English](THIRD_PARTY_NOTICES.en.md)

## 2.1 在线版新增组件（2026-09-18）

- `@ffmpeg/ffmpeg 0.12.15`：MIT JavaScript Worker 包装器，版权归 Jerome Wu 与 ffmpeg.wasm 贡献者。
- `@ffmpeg/core 0.12.10`：**GPL-2.0-or-later** 的 WebAssembly 引擎，不属于本项目 MIT 自有代码。在线站点按原样分发该二进制，不加载第三方 CDN。GPL 全文见 `licenses/FFmpeg-GPL-2.0.txt`，在线副本为 `/FFmpeg-GPL-2.0.txt`。
- 上游源码和构建配方：https://github.com/ffmpegwasm/ffmpeg.wasm/tree/v12.15 （其 packages/core/package.json 对应 0.12.10）。FFmpeg 基础为 n5.1.4，不能将桌面版的新 FFmpeg 测试结论套用到这个引擎。
- 同时提供源代码及构建依赖源码档案： https://github.com/Yangjunjie-Lin/PrismTranscode/releases/download/v2.1.0-beta.1/FFmpeg-WASM-source-bundle.tar 。各源码档案内保留上游许可证，清单记录下载地址与 SHA-256；生成脚本为 `scripts/bundle-wasm-sources.mjs`。
- 引擎包含 x264/x265、libvpx、LAME、Ogg/Vorbis/Theora、Opus、zlib、WebP、FreeType、FriBidi、HarfBuzz、libass、zimg 和 Emscripten 运行时等；具体构建声明以源代码 Dockerfile、各项目许可及二进制 `-version` 输出为准。
- Vite（MIT）、Playwright（Apache-2.0）只用于开发构建/测试；完整锁定依赖见 package-lock.json。依赖审计不是原生编解码器漏洞审计。
- 在线版自有代码仍按 MIT 提供。使用和再分发整个组合时，也必须遵守引擎与链接库的适用 GPL/其他许可，不能声称所有内容均为 MIT。尚未进行独立法律或编解码器专利审查。

以下为 **2.0 桌面包** 的历史说明；其中“未附引擎”仅指 Windows 桌面包，不适用于新增在线站点。

核对日期：2026-09-16。这里明确区分“本项目实际使用的外部组件”和“仅参考的工作流”，不表示第三方作者背书或合作。

| 来源 | 在本项目中的关系 | 许可与说明 |
| --- | --- | --- |
| 原 NCM-MP3 Batch v1.0.0 | 沿用其 Go NCM 解析模块、基础平台适配和合成测试素材，新增媒体管线与界面 | 保留原 MIT 声明 |
| Go | 本项目编译器、标准库与运行时；无第三方 Go modules | BSD 风格；附 `licenses/Go-LICENSE.txt` |
| FFmpeg / FFprobe | 实际调用的外部进程，处理探测、编解码、滤镜与封装 | 许可受构建配置影响；本包未附引擎二进制 |
| BtbN FFmpeg-Builds | 用户可选完整 GPL 构建下载渠道 | 第三方构建；自带许可、来源与构建配置；滚动 latest 不是固定版本 |
| Gyan FFmpeg builds | 用户可选 release essentials 下载渠道 | 第三方构建；功能集合与完整构建不同 |
| SoXR、SVT-AV1、x264/x265、libvpx、LAME、Opus、libaom、libjxl 等 | 由安装的 FFmpeg 构建提供，按可用能力调用 | 不把这些库当成本项目自写；自身许可仍适用 |
| taurusxin/ncmdump | NCM 结构/协议实现参考 | MIT；本项目未调用其可执行文件 |
| mifi/lossless-cut | 参考原码提取与重封装的工作流 | 未合并其 GUI 或发行二进制，非代码再许可 |
| paulpacifico/shutter-encoder | 参考多媒体批处理工作台与功能组织 | 未合并其 GUI 或发行二进制，非代码再许可 |

## 官方/作者资料

- FFmpeg 源码与发行说明：https://ffmpeg.org/download.html
- FFmpeg 命令与 streamcopy：https://ffmpeg.org/ffmpeg.html
- FFprobe 探测：https://ffmpeg.org/ffprobe.html
- FFmpeg 编码器：https://ffmpeg.org/ffmpeg-codecs.html
- 重采样与 SoXR precision：https://ffmpeg.org/ffmpeg-resampler.html
- FFmpeg 许可：https://ffmpeg.org/legal.html
- FFmpeg 源码镜像：https://github.com/FFmpeg/FFmpeg
- BtbN 构建：https://github.com/BtbN/FFmpeg-Builds
- Gyan 构建：https://www.gyan.dev/ffmpeg/builds/
- NCM 参考：https://github.com/taurusxin/ncmdump
- LosslessCut：https://github.com/mifi/lossless-cut
- Shutter Encoder：https://github.com/paulpacifico/shutter-encoder
- Go 发行与安全维护：https://go.dev/dl/ ；https://go.dev/security/

核对时 FFmpeg 官网列出稳定版 9.0.1（2026-08-12 发布）。本次本地真实测试使用 7.1.5，不能把源码官网的最新版本号写成“已在本项目实测”。完整下载器选择的是 BtbN 开发构建，可能比稳定版更新，也可能出现新回归；正式部署应固定经验证的引擎。

FFmpeg 官网提供源码并列出第三方 Windows 构建链接，不代表本软件自带“FFmpeg 官方 Windows 二进制”。本安装器只保存下载渠道、包哈希、时间和程序/许可内容；二次分发引擎时应自行完成相应源码、通知和许可义务。MIT 仅覆盖本项目有权如此许可的自有代码，不覆盖外部 GPL 或其他组件。

本交付测试素材为合成正弦音、测试图形、透明色块与自行写的短字幕；用户提供的歌曲只用于本地回归，不分发在项目源码或示例包中。
