# Meridian Rule Sets

自动生成的代理规则集，支持多种格式。

## 可用规则文件

### GFW 规则集（完整版）

包含 GFW 域名列表 + 本地自定义域名 + Telegram IP 段

| 格式 | GitHub Raw | jsDelivr CDN |
|------|-----------|--------------|
| Plain Text | [gfw.txt](https://raw.githubusercontent.com/soarwind/meridian/release/gfw.txt) | [gfw.txt](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/gfw.txt) |
| YAML (Mihomo) | [gfw.yaml](https://raw.githubusercontent.com/soarwind/meridian/release/gfw.yaml) | [gfw.yaml](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/gfw.yaml) |
| MRS (Mihomo Binary) | [gfw.mrs](https://raw.githubusercontent.com/soarwind/meridian/release/gfw.mrs) | [gfw.mrs](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/gfw.mrs) |
| SRS (Sing-box Binary) | [gfw.srs](https://raw.githubusercontent.com/soarwind/meridian/release/gfw.srs) | [gfw.srs](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/gfw.srs) |

### GFW-Lite 规则集（精简版）

包含 GFW-Lite 域名列表 + 本地自定义域名 + Telegram IP 段

| 格式 | GitHub Raw | jsDelivr CDN |
|------|-----------|--------------|
| Plain Text | [gfw-lite.txt](https://raw.githubusercontent.com/soarwind/meridian/release/gfw-lite.txt) | [gfw-lite.txt](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/gfw-lite.txt) |
| YAML (Mihomo) | [gfw-lite.yaml](https://raw.githubusercontent.com/soarwind/meridian/release/gfw-lite.yaml) | [gfw-lite.yaml](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/gfw-lite.yaml) |
| MRS (Mihomo Binary) | [gfw-lite.mrs](https://raw.githubusercontent.com/soarwind/meridian/release/gfw-lite.mrs) | [gfw-lite.mrs](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/gfw-lite.mrs) |
| SRS (Sing-box Binary) | [gfw-lite.srs](https://raw.githubusercontent.com/soarwind/meridian/release/gfw-lite.srs) | [gfw-lite.srs](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/gfw-lite.srs) |

### IP CIDR 规则

Telegram、WhatsApp、Google Voice 等服务的 IP CIDR 列表

| 格式 | GitHub Raw | jsDelivr CDN |
|------|-----------|--------------|
| Plain Text | [ipcidr.txt](https://raw.githubusercontent.com/soarwind/meridian/release/ipcidr.txt) | [ipcidr.txt](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/ipcidr.txt) |
| YAML | [ipcidr.yaml](https://raw.githubusercontent.com/soarwind/meridian/release/ipcidr.yaml) | [ipcidr.yaml](https://cdn.jsdelivr.net/gh/soarwind/meridian@release/ipcidr.yaml) |

## 更新频率

- **自动更新**: 每天 UTC 00:00 (北京时间 08:00)
- **触发更新**: 当 `domain.txt` 或 `ipcidr.txt` 有变更时
- **手动更新**: 可在 GitHub Actions 手动触发

## 使用说明

### Mihomo (Clash Meta)

```yaml
rule-providers:
  gfw:
    type: http
    behavior: domain
    format: mrs
    url: "https://raw.githubusercontent.com/soarwind/meridian/release/gfw.mrs"
    path: ./ruleset/gfw.mrs
    interval: 86400

rules:
  - RULE-SET,gfw,PROXY
```

### Sing-box

```json
{
  "route": {
    "rule_set": [
      {
        "type": "remote",
        "tag": "gfw",
        "format": "binary",
        "url": "https://raw.githubusercontent.com/soarwind/meridian/release/gfw.srs",
        "download_detour": "proxy"
      }
    ],
    "rules": [
      {
        "rule_set": ["gfw"],
        "outbound": "proxy"
      }
    ]
  }
}
```

## 数据来源

- **GFW 域名**: [Loyalsoldier/surge-rules](https://github.com/Loyalsoldier/surge-rules)
- **GFW-Lite 域名**: [Loyalsoldier/surge-rules](https://github.com/Loyalsoldier/surge-rules)
- **Telegram IP**: [Loyalsoldier/surge-rules](https://github.com/Loyalsoldier/surge-rules)
- **本地自定义**: 仓库中的 `domain.txt` 和 `ipcidr.txt`

## 源代码

查看源代码和自定义规则: [soarwind/meridian (main 分支)](https://github.com/soarwind/meridian)
