# Outline Hexo Connector

一个用于自动同步 [Outline](https://www.getoutline.com/) 文档到 [Hexo](https://hexo.io/) 博客的 Webhook 处理器。

[English](README.md)

## 📝 简介

Outline Hexo Connector 是一个轻量级的 Go 服务，用于监听 Outline Wiki 的 Webhook 事件，并自动将文档内容同步到 Hexo 静态博客系统。当 Outline 中的文档发生变化时（如创建、发布、更新或删除），本服务会自动处理这些事件并触发相应的操作。

## ✨ 特性

- 🔐 **安全验证**：支持 Outline Webhook 签名验证，确保请求来源的可靠性
- 📋 **事件处理**：支持多种文档事件（创建、发布、取消发布、归档、删除等）
- 🧪 **测试模式**：内置测试模式，方便调试 Webhook 请求
- ⚙️ **灵活配置**：通过 YAML 配置文件管理所有设置
- 🔍 **集合过滤**：可指定特定的 Outline 集合用于博客发布
- 🌐 **RESTful API**：与 Outline API 完整集成
- 🎯 **附件处理**：支持获取附件的重定向 URL

## 🚀 快速开始

### 前置要求

- Go 1.21 或更高版本
- 运行中的 Outline 实例
- Hexo 博客项目（即将支持）

### 安装

```bash
# 克隆仓库
git clone https://github.com/Charles-IX/outline-hexo-connector.git
cd outline-hexo-connector

# 安装依赖
go mod download

# 构建
go build -o outline-webhook
```

## ⚙️ 配置

1. 复制配置示例文件：

```bash
cp config_example.yaml config.yaml
```

2. 编辑 `config.yaml` 并填写你的配置信息：

```yaml
# Outline API 密钥
Outline_API_Key: your_api_key_here

# Outline API 地址
Outline_API_URL: https://outline.example.com/api

# Webhook 密钥（用于验证请求签名）
Outline_Webhook_Secret: your_webhook_secret_here

# 用于博客发布的集合名称
Outline_Collection_Used_For_Blog: Blog

# Hexo 构建超时时间（秒）
Hexo_Build_Timeout: 30
```

### 配置说明

| 配置项 | 说明 | 必填 |
|-------|------|------|
| `Outline_API_Key` | Outline API 访问密钥 | ✅ |
| `Outline_API_URL` | Outline API 端点地址 | ✅ |
| `Outline_Webhook_Secret` | Webhook 签名验证密钥 | ✅ |
| `Outline_Collection_Used_For_Blog` | 指定用于博客的集合名称 | ✅ |
| `Hexo_Build_Timeout` | Hexo 构建超时时间（秒） | ✅ |

## 📖 使用方法

### 启动服务

默认启动（使用 `config.yaml` 配置文件，监听 9000 端口）：

```bash
./outline-webhook
```

### 命令行参数

```bash
./outline-webhook [OPTIONS]
```

**可用选项：**

- `-p, --port <port>`：指定监听端口（默认：9000）
- `-c, --config <path>`：指定配置文件路径（默认：config.yaml）
- `-t, --test`：启用测试模式，仅打印接收到的原始请求

### 示例

```bash
# 使用自定义端口
./outline-webhook -p 8080

# 使用自定义配置文件
./outline-webhook -c /path/to/config.yaml

# 启用测试模式
./outline-webhook -t

# 组合使用
./outline-webhook -p 8080 -c custom.yaml
```

### 配置 Outline Webhook

1. 登录你的 Outline 管理面板
2. 进入 **Settings** → **API & Webhooks**
3. 创建新的 Webhook：
   - **URL**: `http://your-server:9000/webhook`
   - **Secret**: 与 `config.yaml` 中的 `Outline_Webhook_Secret` 保持一致
   - **Events**: 选择需要监听的事件类型

## 📦 项目结构

```
outline-webhook/
├── main.go                 # 主程序入口
├── config_example.yaml     # 配置示例（使用时应改名为config.yaml）
├── go.mod                  # Go 模块定义
├── README.md               # 项目文档
└── internal/
    ├── config/
    │   └── config.go       # 配置管理
    ├── outline/
    │   ├── client.go       # Outline API 客户端
    │   └── models.go       # 数据模型
    ├── hexo/
    │   └── adapter.go      # Hexo 适配器（开发中）
    ├── processor/
    │   ├── converter.go    # 内容转换器（开发中）
    │   └── markdown.go     # Markdown 处理（开发中）
    └── test/
        └── test.go         # 测试工具
```

## 🔍 支持的事件类型

| 事件类型 | 说明 | 状态 |
|---------|------|------|
| `documents.create` | 文档创建 | 🚧 开发中 |
| `documents.publish` | 文档发布 | 🚧 开发中 |
| `documents.update` | 文档更新 | 🚧 开发中 |
| `documents.unpublish` | 取消发布 | 🚧 开发中 |
| `documents.archive` | 文档归档 | 🚧 开发中 |
| `documents.unarchive` | 取消归档 | 🚧 开发中 |
| `documents.restore` | 文档恢复 | 🚧 开发中 |
| `documents.delete` | 文档删除 | 🚧 开发中 |
| `documents.move` | 文档移动 | 🚧 开发中 |
| `documents.title_change` | 标题更改 | 🚧 开发中 |

## 🛠️ 开发

### 依赖项

- [pflag](https://github.com/spf13/pflag) - 命令行参数解析
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML 配置文件解析

### 运行测试模式

测试模式允许你查看接收到的原始 Webhook 请求：

```bash
./outline-webhook -t
```

然后从 Outline 触发一个测试事件，你将在控制台看到完整的请求内容。

## 📋 待办事项

- [ ] 完善 Hexo 适配器实现
- [ ] 实现文档到 Markdown 的完整转换
- [ ] 添加附件 URL 转换功能
- [ ] 实现文档发布/删除时的 Hexo 构建触发
- [ ] 添加文档队列机制，支持定期批量构建
- [ ] 添加单元测试
- [ ] 完善错误处理和日志记录
- [ ] 支持数据库存储文档映射关系（存疑）
- [ ] 添加 Docker 支持

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
~~不过项目基本功能还没完全实现，真的会有人交么~~

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [Outline](https://www.getoutline.com/) - 强大的团队知识库
- [Hexo](https://hexo.io/) - 快速、简洁的博客框架

## 📞 联系方式

如有问题或建议,请通过以下方式联系：

- GitHub Issues: [https://github.com/Charles-IX/outline-hexo-connector/issues](https://github.com/Charles-IX/outline-hexo-connector/issues)

---

⚠️ **注意**：本项目目前处于活跃开发阶段，部分功能尚未完成。不建议~~在生产环境中~~使用。
