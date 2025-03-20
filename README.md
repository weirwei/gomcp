# Gomcp

Gomcp 是一个用 Go 语言编写的 MCP (Message Control Protocol) 客户端和服务器的 SDK。

## 功能特性

- 支持 Unix 系统下的进程间通信
- 支持标准输入输出（stdio）模式

## 安装

```bash
go get github.com/weirwei/gomcp
```

## 客户端配置

创建配置文件 `.mcp-config.json`，参考以下示例：

```json
{
    "mcpServers": {
        "server-name": {
            "command": "command-to-run",
            "args": ["arg1", "arg2"],
            "env": {
                "ENV_VAR": "value"
            },
            "disabled": false
        }
    }
}
```

### 配置项说明

- `command`: 要执行的命令
- `args`: 命令参数数组
- `env`: 环境变量配置
- `disabled`: 是否禁用该服务器

## 使用方法

查看 `examples` 目录获取使用示例。
