# 本机 API（prismtranscode 2.0.0）

供本机前端与测试使用，不是公网接口。主程序 stdout 和数据目录 last-session-url.txt 提供 `http://127.0.0.1:<port>/app/<token>/`。API 使用 `Authorization: Bearer <token>`；请求 Host 必须与启动地址相同，存在 Origin 时必须同源。读取与写入都鉴权，写入要求 POST。令牌每次启动重新生成。

## 主要端点

| 方法 | 路径 | 请求 / 响应 |
| --- | --- | --- |
| GET | `/api/config` | 版本、平台、引擎、能力、输出预设、完整 defaults |
| GET | `/api/queue` | jobs、busy、output_dir、workers |
| POST | `/api/add` | `{paths:[],folder:"",recursive:false,options:{...}}`；添加普通路径，可混合文件夹 |
| POST | `/api/upload` | multipart：`options` JSON 字符串与 `files`；在本机生成 imports 副本 |
| POST | `/api/configure` | `{ids:[],options:{...}}`，Options 为完整对象，不是局部 patch |
| POST | `/api/remove` | `{ids:[]}`，只移除任务，不删除源或成品 |
| POST | `/api/start` | `{ids:[],output_dir:"...",workers:1}` |
| POST | `/api/stop` | `{}`，取消当前队列 |
| POST | `/api/plan` | `{id:"..."}`，返回同一编译逻辑生成的计划 |
| GET | `/api/export?format=json` | 完整任务快照，包含日志、警告与输出信息 |
| GET | `/api/export?format=csv` | 带 UTF-8 BOM 的简表；文本字段防公式注入 |
| POST | `/api/engine` | `{path:"...",browse:false}` 或 `{browse:true}`；更换引擎前不得有运行任务 |
| POST | `/api/install` | `{source:"full"}` 或 `{source:"essentials"}`；仅 Windows |
| GET | `/api/install-status` | busy、phase、percent、error |
| POST | `/api/cancel-install` | `{}` |
| POST | `/api/browse-files` | `{}`，Windows 原生文件多选 |
| POST | `/api/browse-folder` | `{}`，Windows 文件夹选择 |
| POST | `/api/open-output` | `{path:"..."}`，只打开已经存在的本地目录 |
| POST | `/api/shutdown` | `{}`，停止任务并退出 |

Options 字段和严格范围以 `internal/media/plan.go` 为准。先从 `/api/config` 读取完整 defaults，再覆盖需要修改的字段，不要手工遗漏布尔字段或线程数。返回错误是 JSON `{error:"..."}`，HTTP 400/403/405 等，不是每次都 200。

示意代码（仅用于自己启动的本机应用）：

```python
import json
import urllib.request

def request(base, token, path, body=None):
    data = None if body is None else json.dumps(body).encode("utf-8")
    headers = {"Authorization": "Bearer " + token}
    if data is not None:
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(base + "/api/" + path, data=data, headers=headers)
    with urllib.request.urlopen(req, timeout=90) as response:
        return json.load(response)

# base/token 从本次本机启动地址读取；不要硬编码或分享。
# config = request(base, token, "config")
# opts = {**config["defaults"], "target": "flac"}
# request(base, token, "add", {"paths": [r"D:\Music\input.wav"], "options": opts})
```

普通 JSON 请求限 2 MiB；一次浏览器导入限制 8 GiB、200 个文件，队列总数最多 5000。直接路径不受 8 GiB 上传限制，但仍受本地磁盘、内存、编解码器和文件系统能力制约。为避免同名冲突只允许编号或跳过，不提供覆盖标志。
