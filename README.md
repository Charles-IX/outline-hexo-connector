# Outline Hexo Connector

A Webhook handler for automatically syncing [Outline](https://www.getoutline.com/) documents to [Hexo](https://hexo.io/) blog.

[中文](README_zh.md)

## 📝 Introduction

Outline Hexo Connector is a lightweight Go service that listens to Outline Wiki's Webhook events and automatically synchronizes document content to the Hexo static blog system. When documents in Outline change (such as create, publish, update, or delete), this service automatically handles these events and triggers corresponding actions.

## ✨ Features

- 🔐 **Security Verification**: Supports Outline Webhook signature verification to ensure request authenticity
- 📋 **Event Handling**: Supports multiple document events (create, publish, unpublish, archive, delete, etc.)
- 🧪 **Test Mode**: Built-in test mode for debugging Webhook requests
- ⚙️ **Flexible Configuration**: Manage all settings through YAML configuration files
- 🔍 **Collection Filtering**: Specify particular Outline collections for blog publishing
- 🌐 **RESTful API**: Full integration with Outline API
- 🎯 **Attachment Handling**: Support for fetching attachment redirect URLs

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- Running Outline instance
- Hexo blog project (coming soon)

### Installation

```bash
# Clone the repository
git clone https://github.com/Charles-IX/outline-hexo-connector.git
cd outline-hexo-connector

# Install dependencies
go mod download

# Build
go build -o outline-webhook
```

## ⚙️ Configuration

1. Copy the example configuration file:

```bash
cp config_example.yaml config.yaml
```

2. Edit `config.yaml` and fill in your configuration information:

```yaml
# Outline API Key
Outline_API_Key: your_api_key_here

# Outline API URL
Outline_API_URL: https://outline.example.com/api

# Webhook Secret (for verifying request signatures)
Outline_Webhook_Secret: your_webhook_secret_here

# Collection name used for blog publishing
Outline_Collection_Used_For_Blog: Blog

# Hexo build timeout (seconds)
Hexo_Build_Timeout: 30
```

### Configuration Reference

| Configuration Item | Description | Required |
|-------------------|-------------|----------|
| `Outline_API_Key` | Outline API access key | ✅ |
| `Outline_API_URL` | Outline API endpoint URL | ✅ |
| `Outline_Webhook_Secret` | Webhook signature verification secret | ✅ |
| `Outline_Collection_Used_For_Blog` | Specify collection name for blog | ✅ |
| `Hexo_Build_Timeout` | Hexo build timeout (seconds) | ✅ |

## 📖 Usage

### Starting the Service

Default start (using `config.yaml` configuration file, listening on port 9000):

```bash
./outline-webhook
```

### Command Line Arguments

```bash
./outline-webhook [OPTIONS]
```

**Available options:**

- `-p, --port <port>`: Specify listening port (default: 9000)
- `-c, --config <path>`: Specify configuration file path (default: config.yaml)
- `-t, --test`: Enable test mode, only print raw incoming requests

### Examples

```bash
# Use custom port
./outline-webhook -p 8080

# Use custom configuration file
./outline-webhook -c /path/to/config.yaml

# Enable test mode
./outline-webhook -t

# Combined usage
./outline-webhook -p 8080 -c custom.yaml
```

### Configuring Outline Webhook

1. Log in to your Outline admin panel
2. Navigate to **Settings** → **API & Webhooks**
3. Create a new Webhook:
   - **URL**: `http://your-server:9000/webhook`
   - **Secret**: Keep consistent with `Outline_Webhook_Secret` in `config.yaml`
   - **Events**: Select the event types you want to listen to

## 📦 Project Structure

```
outline-webhook/
├── main.go                 # Main program entry
├── config_example.yaml     # Configuration example (rename to config.yaml when using)
├── go.mod                  # Go module definition
├── README.md               # Project documentation
└── internal/
    ├── config/
    │   └── config.go       # Configuration management
    ├── outline/
    │   ├── client.go       # Outline API client
    │   └── models.go       # Data models
    ├── hexo/
    │   └── adapter.go      # Hexo adapter (in development)
    ├── processor/
    │   ├── converter.go    # Content converter (in development)
    │   └── markdown.go     # Markdown processing (in development)
    └── test/
        └── test.go         # Testing utilities
```

## 🔍 Supported Event Types

| Event Type | Description | Status |
|-----------|-------------|--------|
| `documents.create` | Document creation | 🚧 In Development |
| `documents.publish` | Document publication | 🚧 In Development |
| `documents.update` | Document update | 🚧 In Development |
| `documents.unpublish` | Unpublish document | 🚧 In Development |
| `documents.archive` | Document archiving | 🚧 In Development |
| `documents.unarchive` | Unarchive document | 🚧 In Development |
| `documents.restore` | Document restoration | 🚧 In Development |
| `documents.delete` | Document deletion | 🚧 In Development |
| `documents.move` | Document moving | 🚧 In Development |
| `documents.title_change` | Title change | 🚧 In Development |

## 🛠️ Development

### Dependencies

- [pflag](https://github.com/spf13/pflag) - Command-line argument parsing
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML configuration file parsing

### Running Test Mode

Test mode allows you to view raw Webhook requests received:

```bash
./outline-webhook -t
```

Then trigger a test event from Outline, and you will see the complete request content in the console.

## 📋 TODO

- [ ] Complete Hexo adapter implementation
- [ ] Implement full document-to-Markdown conversion
- [x] Add attachment URL conversion functionality (convert from Outline API URL to OSS permanent URL)
- [ ] Implement Hexo build triggering on document publish/delete
- [ ] Add document queue mechanism for periodic batch builds
- [ ] Add unit tests
- [ ] Improve error handling and logging
- [ ] Support database storage for document mapping relationships (uncertain)
- [ ] Add Docker support

## 🤝 Contributing

Issues and Pull Requests are welcome!
~~Though the basic functionality isn't fully implemented yet, would anyone really contribute?~~

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Outline](https://www.getoutline.com/) - Powerful team knowledge base
- [Hexo](https://hexo.io/) - Fast, simple & powerful blog framework

## 📞 Contact

For questions or suggestions, please contact via:

- GitHub Issues: [https://github.com/Charles-IX/outline-hexo-connector/issues](https://github.com/Charles-IX/outline-hexo-connector/issues)

---

⚠️ **Notice**: This project is currently under active development and some features are not yet complete. ~~Not recommended for production use.~~
