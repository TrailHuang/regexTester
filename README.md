# 正则表达式测试工具

一个用于测试正则表达式匹配规则的 Go 语言工具。

## 功能特性

- 支持从 JSON 配置文件加载正则规则
- 支持批量测试和单 URL 测试
- 支持多种匹配方法（regex、not_regex）
- 命令行参数支持
- 跨平台编译

## 快速开始

### 编译程序

```bash
make build
```

### 运行测试

```bash
# 使用默认配置
./build/bin/regex-tester

# 测试单个 URL
./build/bin/regex-tester -u "/api/v1/users?access_token=abc123"

# 使用自定义配置
./build/bin/regex-tester -c custom_config.json -t custom_test.txt
```

### 打包发布

```bash
make package
```

## 配置文件格式

配置文件为 JSON 格式，支持多个规则集：

```json
[
    {
        "name": "api_rules",
        "rules": [
            {
                "method": "regex",
                "value": "(/|(/[^/]*?[^a-zA-Z]+))(api|rest|restful)(([^a-zA-Z]+[^/]*/)|/)"
            },
            {
                "method": "not_regex",
                "value": "(?i)(?:[?&])(access[_-]?token)=([^&?#\\s]+)"
            }
        ]
    }
]
```

## 测试用例格式

测试用例文件为文本格式，包含两列：

```
# URL	期望结果
/api/users	匹配
/api/v1/users?access_token=abc123	不匹配
```

## Makefile 目标

- `make build` - 编译当前平台
- `make build-all` - 编译所有平台
- `make package` - 打包成 tar 包
- `make test` - 运行测试
- `make clean` - 清理构建文件
- `make help` - 显示帮助信息

## 许可证

MIT License