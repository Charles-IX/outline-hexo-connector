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
- Hexo 博客项目（测试时使用了fluid主题）

### 安装

```bash
# 克隆仓库
git clone https://github.com/Charles-IX/outline-hexo-connector.git
cd outline-hexo-connector

# 安装依赖
go mod download

# 构建
go build -o outline-hexo-connector
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

# Hexo 构建触发间隔（秒），防止频繁触发
Hexo_Build_Interval: 30

# Hexo 构建命令
Hexo_Build_Command: hexo clean && hexo generate

# Hexo 文章存放目录（用于写入同步的 Markdown 文件）
Hexo_Source_Post_Dir: hexo/source/_posts
```

### 配置说明

| 配置项 | 说明 | 必填 |
|-------|------|------|
| `Outline_API_Key` | Outline API 访问密钥 | ✅ |
| `Outline_API_URL` | Outline API 端点地址 | ✅ |
| `Outline_Webhook_Secret` | Webhook 签名验证密钥 | ✅ |
| `Outline_Collection_Used_For_Blog` | 指定用于博客的集合名称 | ✅ |
| `Hexo_Build_Interval` | Hexo 构建触发的最小间隔时间（秒），用于防抖 | ✅ |
| `Hexo_Build_Command` | 执行 Hexo 构建的 Shell 命令 | ✅ |
| `Hexo_Source_Post_Dir` | Hexo 博客的 `source/_posts` 目录路径 | ✅ |

### 支持的事件类型

Connector 目前支持监听并处理以下 Outline Webhook 事件：

- **发布与更新事件** (触发文章创建/更新 + Hexo 构建):
    - `documents.publish`: 文档发布时
    - `documents.unarchive`: 文档从归档恢复时
    - `documents.restore`: 文档从回收站恢复时
    - `documents.move`: 文档移动时 (更新分类信息)
    - `documents.title_change`: 文档标题变更时
    - `documents.update`: 文档内容更新时

- **删除与归档事件** (触发文章删除 + Hexo 构建):
    - `documents.unpublish`: 文档取消发布时
    - `documents.archive`: 文档被归档时
    - `documents.delete`: 文档被删除时

- **其他事件**:
    - `documents.create`: 仅做内部逻辑处理，防止草稿被意外发布

## 📖 使用方法

### 启动服务

默认启动（使用 `config.yaml` 配置文件，监听 9000 端口）：

```bash
./outline-hexo-connector
```

### 命令行参数

```bash
./outline-hexo-connector [OPTIONS]
```

**可用选项：**

- `-p, --port <port>`：指定监听端口（默认：9000）
- `-c, --config <path>`：指定配置文件路径（默认：config.yaml）
- `-t, --test`：启用测试模式，仅打印接收到的原始请求

### 示例

```bash
# 使用自定义端口
./outline-hexo-connector -p 8080

# 使用自定义配置文件
./outline-hexo-connector -c /path/to/config.yaml

# 启用测试模式
./outline-hexo-connector -t

# 组合使用
./outline-hexo-connector -p 8080 -c custom.yaml
```

### 配置 Outline Webhook

1. 登录你的 Outline 管理面板
2. 进入 **偏好设置** → **Webhooks**
3. 创建新的 Webhook：
   - **URL**: `http://Outline-Hexo-Connector的IP:端口/webhook`
   - **Secret**: 复制到 `config.yaml` 中的 `Outline_Webhook_Secret`
   - **Events**: 选择需要监听的事件类型，建议包含 `documents.create`, `documents.publish`, `documents.unpublish`, `documents.delete`, `documents.archive`, `documents.unarchive`, `documents.restore`, `documents.move`, `documents.update`, `documents.title_change`
4. 进入 **偏好设置** → **API 与应用程序**
5. 创建新的 API 密钥：
   - **作用域**: 至少需要 `documents.info`, `documents.unpublish`, `collections.info`, `attachments.redirect`
   - **过期时间**: 根据自己需求而定
   - 将创建好的 API 密钥复制到 `config.yaml` 中的 `Outline_API_Key`

## ⚠️ 说明

由于Outline会自动发布刚刚创建好的新文档，为了避免在Hexo中新建一个无意义的空文件并触发构建，本工具会自动将刚创建好的文档取消发布，待编辑完成后再次发布即可。

本工具也会自动将作用范围内的有更新的文档取消发布，以便用户可以通过点击“发布”来构建Hexo博客。

## 🏷️ 文档自定义标签指南

为了让同步到 Hexo 的文章具备完整的元数据（如标签、摘要、封面图），本工具支持了一套自定义的 Markdown 语法标签。这些标签在同步过程中会被解析处理，不会直接显示在文章正文中。

### 1. 文章标签 (Tags)

用于为 Hexo 文章设置标签。支持使用英文逗号 `,` 或中文逗号 `，` 分隔。

- **语法**：`+> Tags: tag1, tag2`
- **位置**：建议放在文档开头或结尾。
- **效果**：解析为 Front Matter 中的 `tags: [tag1, tag2]`，并从正文中移除。

### 2. 摘要分隔符 (Read More)

控制文章在首页列表中的摘要显示范围。

- **语法**：`+> More:`
- **效果**：替换为 Hexo 的 `<!-- more -->` 标记。在此标记之前的内容将作为摘要显示。

### 3.封面与缩略图 (Banner & Index Image)

设置文章的顶部大图（Banner）和列表缩略图（Index Image）。语法类似于标准的 Markdown 图片语法，但使用特定的 Alt 文本。

| 语法 | 说明 |
|------|------|
| `![banner_img](url)` | 仅设置文章详情页的顶部封面图 (Banner) |
| `![index_img](url)` | 仅设置文章列表页的缩略图 (Index) |
| `![banner_index_img](url)` | 同时设置 Banner 和 Index 图片 |
| `![index_banner_img](url)` | 同上，同时设置 Banner 和 Index 图片 |

> **注意**：这些特殊的图片标签在解析后会从正文中移除，转化为 Front Matter 配置。

### 示例

在 Outline 文档中：

```markdown
# 我的新文章

+> Tags: 技术, Golang, 教程

这是文章的摘要部分。

+> More:

![banner_index_img](https://example.com/cover.jpg)

这里是文章的正文内容...
```

## 📦 项目结构

```
outline-hexo-connector/
├── main.go                 # 主程序入口，处理命令行参数和信号
├── config_example.yaml     # 配置示例文件
├── go.mod                  # Go 模块定义
├── README.md               # 英文文档
├── README_zh.md            # 中文文档
└── internal/
    ├── config/
    │   └── config.go       # 配置加载与解析
    ├── hexo/
    │   ├── renderer.go     # Hexo 文章文件生成与写入
    │   └── trigger.go      # Hexo 构建命令触发与防抖控制
    ├── outline/
    │   ├── client.go       # Outline API 客户端与 Webhook 处理
    │   └── models.go       # Outline 数据模型定义
    ├── processor/
    │   ├── converter.go    # 附件 URL 转换与处理
    │   └── parser.go       # Markdown 内容解析与元数据提取
    └── test/
        └── test.go         # 测试工具与 Debug 辅助
```

## 🛠️ 开发

### 依赖项

- [pflag](https://github.com/spf13/pflag) - 命令行参数解析
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML 配置文件解析

### 运行测试模式

测试模式允许你查看接收到的原始 Webhook 请求：

```bash
./outline-hexo-connector -t
```

然后从 Outline 触发一个测试事件，你将在控制台看到完整的请求内容。

## 📋 待办事项

- [x] 完善 Hexo 适配器实现
- [x] 实现文档到 Markdown 的完整转换
- [x] 添加附件 URL 转换功能（从Outline API转到OSS永久链接）
- [x] 实现文档发布/删除时的 Hexo 构建触发
- [x] 添加文档队列机制，支持定期批量构建
- [ ] 添加单元测试（也许）
- [x] 完善错误处理和日志记录

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
这是一个我出于实用性考量和练习Go而写的小程序，请多指教。
~~不过项目基本功能还没完全实现，真的会有人交么~~
项目基本功能已经实现，但是还有一些限制。之后可能会逐步完善。

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [Outline](https://www.getoutline.com/) - 强大的团队知识库
- [Hexo](https://hexo.io/) - 快速、简洁的博客框架

## 📞 联系方式

如有问题或建议,请通过以下方式联系：

- GitHub Issues: [https://github.com/Charles-IX/outline-hexo-connector/issues](https://github.com/Charles-IX/outline-hexo-connector/issues)

---
