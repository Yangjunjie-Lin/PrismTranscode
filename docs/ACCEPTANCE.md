# 验收记录与发布门槛

## 已验证的内容

本次核心环境：Linux、FFmpeg 7.1.5-0+deb13u1、Go 1.23.2。工具链具体版本、程序与源文件哈希见交付根目录 release-manifest.json。测试数据以合成媒体为主，另对用户提供的一份 NCM 做本地回归。

**单元测试：33 个顶层测试通过，包含 race 检查；go vet 通过。** 覆盖 NCM 长度/签名/取消、参数白名单、HDR 防误转换、MP4 重封装兼容性、硬件回退规划、同名排他提交、CSV 公式保护、本机接口鉴权、显式引擎路径、安装包摘要与提取校验、目录扫描、去重和队列恢复。

**真实引擎集成：60/60 检查通过。** 其中 38 个输出预设逐个用适配的真实样例编码、重新探测、完整解码（字幕重新解析）并核对输出 SHA-256。额外验证 24-bit WAV→FLAC PCM 一致，RGBA PNG→PNG/无损 WebP/TIFF 像素一致，MP4→MKV 视频/音频压缩包一致，NCM 原码 MP3 包一致；验证 SoXR、不可用 NVENC 的 CPU 回退、损坏文件隔离、取消清理、无覆盖与跳过策略、AVIF 类型识别和进程重启恢复。

**界面组件：11/11 检查通过。** 使用实际 HTML/CSS/JS，驱动真实本机后端完成 6 个混合媒体任务，验证逐文件格式、Unicode 导入、本地 multipart 导入处理、详情、JSON 导出数据、引擎能力、桌面/窄屏无页面横向溢出和无 JavaScript 异常。

当前环境的 Chromium 阻止 URL 导航，因此界面验证采用内存页面 + Python 本机 API 桥接。它不是 Windows 原生界面验收，也不验证浏览器地址导航、下载提示和系统文件选择器。截图为这条测试链路中的实际程序组件与真实转换结果，不是手绘效果图。正常浏览器端到端脚本 `scripts/ui_test.py` 已提供，尚需在可正常浏览本机页面的开发环境运行。

输出画质精度未进行大规模主观视听或 VMAF/PESQ 测评；不报告虚构的全格式质量评分。60 个检查不意味着所有分辨率、位深、多轨道、元数据、超大文件和损坏样本均已覆盖。

## 必须补齐才能作为正式商业发行的项目

- Windows 10/11 x64 实机启动、默认浏览器、系统文件/文件夹选择器、中文长路径、杀进程与磁盘满、不同文件系统、系统安全提示。
- BtbN/Gyan 联网下载安装的端到端验收，包括重定向、代理、下载取消、哈希不一致和发行更新窗口；本次只完成安装器本地单元测试。
- NVIDIA / Intel / AMD 实际显卡与驱动验证；目前只有软件编码和无硬件条件下的回退实测，不能宣称硬件矩阵已通过。
- 所选新版固定 FFmpeg 构建完整复测；当前最新稳定版 9.0.1 未在本环境执行，不混同于 7.1.5 测试记录。
- 使用当前受支持 Go 工具链重建、漏洞扫描、单实例锁/事务保护方案、原生媒体进程沙箱、代码签名、许可证/专利及媒体授权审查。
- 大文件（超 8 GiB 路径输入）、复杂 HDR/透明/多音轨素材、性能与磁盘预算压力测试。

这些是未完成验收项，不是隐藏在软件界面中的已实现承诺。当前适合作为可运行工程基础、个人本机转换与后续发布验证，不应不经审查直接作为已认证的商业服务上线。

## 复现命令

```sh
go test -race -v ./...
go vet ./...
go build -trimpath -buildvcs=false -o prismtranscode .
python scripts/smoke_test.py --binary ./prismtranscode --work ./tmp-smoke --report ./smoke-report.json
# 原有歌曲不随包分发；可在自己拥有的文件上添加 --sample-ncm 本地路径。
```

集成测试会主动生成和读取短合成媒体，使用本机 FFmpeg，不从网上下载测试歌曲。需要所测的完整编码器能力。`--work` 应使用新的独立目录；旧输出会触发同名保护，可能影响特定断言。

常规浏览器验收：

```sh
python -m pip install playwright
python -m playwright install chromium
python scripts/ui_test.py --binary ./prismtranscode --fixtures ./tmp-smoke/fixtures --work ./tmp-ui --report ./ui-report.json --screenshots ./screenshots
```

CI 配置仅提供流程，未自动创建 GitHub 仓库、提交用户代码或触发远程运行。把 source 目录作为仓库根目录使用即可。
