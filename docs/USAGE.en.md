# Online user guide

[简体中文](USAGE.zh-CN.md) · **English** · [Project overview](../README.en.md)

Open https://prismtranscode.vercel.app . Process only trusted media you have the right to use. No account or native engine installation is required.

## Choosing a language

The top-right selector offers **Browser default / 中文 / English**. On first use, the app matches the browser’s preferred language list against `zh` and `en`, including regional variants. If neither matches, it uses English. Chinese variants all display Simplified Chinese.

A manual choice is saved in this site’s browser `localStorage` under `prismtranscode.language`. Choosing **Browser default** deletes it. When storage is blocked, switching still works, but the manual choice is not remembered after leaving the page. Automatic mode also responds to browser language-change events and checks again on the next visit.

Switching does not reload the page or discard files, per-file formats, quality settings, active jobs or completed downloads. Statuses, application errors, accessibility labels, the page title and documentation links update too. JSON field names and format IDs remain stable; descriptions use the selected language when exported, and `language` records `en` or `zh`. Raw engine diagnostics are not machine-translated.

## Convert files

1. Choose Audio, Video, Images or Subtitles, then the default output format, quality, start/duration and maximum video size.
2. Browse for files or drag them onto the drop zone. New files use the current defaults; each row can also have its own output format.
3. After changing the settings panel, click **Apply to unfinished files** to update existing jobs. Completed jobs remain unchanged; remove and add a file again to reconvert it.
4. Click **Start converting**. The approximately 32 MB engine downloads on first use. Media is read into browser memory, not uploaded. Files run sequentially; a failed file does not abort the batch.
5. Click **Download** for each result. **Export results** saves a JSON report with options, states, output SHA-256, verification descriptions and available failure diagnostics.

**Stop this batch** terminates the active Worker. Starting again retries unfinished jobs and skips completed ones. Removing a job or clearing the queue releases its result. **Refreshing or closing the tab discards all media and results; download them first.**

## Presets and limits

Audio: MP3, WAV (16-bit PCM), FLAC, M4A/AAC, OGG/Vorbis and Opus. Video: MP4/H.264 + AAC and WebM/VP8 + Vorbis. Images: GIF, PNG, JPEG and lossless WebP. Subtitles: SRT and WebVTT.

- At most 128 MiB per file, and 384 MiB / 50 input files per queue. Decoding can use much more memory; mobile devices may run out.
- GIFs are limited to 10 seconds, a maximum width of 480 pixels and 12 fps. Other image presets output only the first frame. JPEG does not preserve transparency.
- Start and duration are in seconds; duration 0 means until the end. Maximum video size downsizes without intentional upscaling. Jobs exceeding 5 minutes are terminated.
- Opus uses FFmpeg’s experimental encoder at 48 kHz stereo. Use desktop libopus for demanding work. Online VP9/libopus is not part of the current validated output configuration.
- Video keeps the first video and audio streams. Complete preservation of attachments, all tracks, private metadata and subtitle styles is not guaranteed.
- HDR-to-image/video conversion is rejected. There is no HDR tone mapping, GPU encoding, NCM decoding or image-subtitle OCR. Use desktop for NCM and large files.
- Verification checks the FFprobe output structure and SHA-256; it is not full-file decoding, visual-quality certification or content-security certification. Lossless formats cannot restore information already lost.

## Troubleshooting

If the engine fails to load, check your connection, browser updates and available memory. Audio outputs need an audio stream; video/images need a picture stream; subtitle outputs need text subtitles. Try a smaller file, a shorter clip or the desktop app. Raw failure diagnostics are available in the JSON report; remove filenames and other private details before sharing it.

Use a current desktop browser; Chromium is the fully tested end-to-end baseline. The host may keep normal access logs for website and engine requests. Manual language preference is the only newly persisted browser setting; media is not stored in localStorage. See [security and privacy](../SECURITY.en.md) and [third-party notices](../THIRD_PARTY_NOTICES.en.md).
