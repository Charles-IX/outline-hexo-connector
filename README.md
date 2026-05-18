# Outline Hexo Connector

A Webhook handler to automatically sync [Outline](https://www.getoutline.com/) documents to [Hexo](https://hexo.io/) blogs.

[中文](README_zh.md)

## 📝 Introduction

Outline Hexo Connector is a lightweight Go service that listens for Outline Wiki Webhook events and automatically syncs document content to the Hexo static blog system. When documents in Outline change (e.g., created, published, updated, or deleted), this service automatically handles these events and triggers the corresponding actions.

## ✨ Features

- 🔐 **Secure Verification**: Supports Outline Webhook signature verification to ensure the reliability of the request source.
- 📋 **Event Handling**: Supports various document events (create, publish, unpublish, archive, delete, etc.).
- 🧪 **Test Mode**: Built-in test mode for easy debugging of Webhook requests.
- ⚙️ **Flexible Configuration**: Manage all settings via a YAML configuration file.
- 🔍 **Collection Filtering**: Specify a specific Outline collection for blog publishing.
- 🌐 **RESTful API**: Fully integrated with the Outline API.
- 🎯 **Attachment Handling**: Supports retrieving redirect URLs for attachments.

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- A running Outline instance
- Hexo blog project (tested with the fluid theme framework)

### Installation

```bash
# Clone the repository
git clone https://github.com/Charles-IX/outline-hexo-connector.git
cd outline-hexo-connector

# Install dependencies
go mod download

# Build
go build -o outline-hexo-connector
```

## ⚙️ Configuration

1. Copy the example configuration file:

```bash
cp config_example.yaml config.yaml
```

2. Edit `config.yaml` and fill in your configuration:

```yaml
# Outline API Key
Outline_API_Key: your_api_key_here

# Outline API URL
Outline_API_URL: https://outline.example.com/api

# Webhook Secret (used to verify request signatures)
Outline_Webhook_Secret: your_webhook_secret_here

# Collection name used for blog publishing
Outline_Collection_Used_For_Blog: Blog

# Hexo build interval (seconds), to prevent frequent triggers
Hexo_Build_Interval: 30

# Hexo build command
Hexo_Build_Command: hexo clean && hexo generate

# Hexo post directory (where synced Markdown files are written)
Hexo_Source_Post_Dir: hexo/source/_posts

# Browser direct-access behavior: html | proxy | disabled
Direct_Access_Mode: html

# Proxy target when Direct_Access_Mode=proxy
Direct_Access_Proxy_Target: http://127.0.0.1:1313

# Custom HTML file when Direct_Access_Mode=html (optional)
Direct_Access_HTML_File: ""
```

### Configuration Details

| Config Item | Description | Required |
|-------------|-------------|----------|
| `Outline_API_Key` | Outline API access key | ✅ |
| `Outline_API_URL` | Outline API endpoint URL | ✅ |
| `Outline_Webhook_Secret` | Webhook signature verification secret | ✅ |
| `Outline_Collection_Used_For_Blog` | Collection name designated for the blog | ✅ |
| `Hexo_Build_Interval` | Minimum interval for Hexo builds (seconds), for debouncing | ✅ |
| `Hexo_Build_Command` | Shell command to execute Hexo build | ✅ |
| `Hexo_Source_Post_Dir` | Path to Hexo blog's `source/_posts` directory | ✅ |
| `Direct_Access_Mode` | Behavior for direct browser access to service port: `html`, `proxy`, or `disabled` | ✅ |
| `Direct_Access_Proxy_Target` | Local target URL when `Direct_Access_Mode=proxy` | ⚠️ Required in proxy mode |
| `Direct_Access_HTML_File` | Optional custom HTML file path when `Direct_Access_Mode=html` | Optional |

### Direct Browser Access

- `POST /webhook`: keeps strict signature verification (for Outline delivery)
- `GET/HEAD /webhook`: returns configured HTML page or reverse-proxies to local target
- `GET /`: same direct-access behavior as above
- Other methods on `/webhook`: returns `405 Method Not Allowed`

### Supported Event Types

The Connector currently supports utilizing the following Outline Webhook events:

- **Publish & Update Events** (Trigger post create/update + Hexo build):
    - `documents.publish`: When a document is published
    - `documents.unarchive`: When a document is restored from the archive
    - `documents.restore`: When a document is restored from the trash
    - `documents.move`: When a document is moved (updates category)
    - `documents.title_change`: When a document title changes
    - `documents.update`: When document content is updated

- **Delete & Archive Events** (Trigger post delete + Hexo build):
    - `documents.unpublish`: When a document is unpublished
    - `documents.archive`: When a document is archived
    - `documents.delete`: When a document is deleted

- **Other Events**:
    - `documents.create`: Internal logic handling only, creates draft to prevent accidental publishing

## 📖 Usage

### Start Service

Default start (uses `config.yaml` and listens on port 9000):

```bash
./outline-hexo-connector
```

### Command Line Arguments

```bash
./outline-hexo-connector [OPTIONS]
```

**Available Options:**

- `-p, --port <port>`: Specify listening port (default: 9000)
- `-c, --config <path>`: Specify config file path (default: config.yaml)
- `-t, --test`: Enable test mode, print raw received requests only

### Examples

```bash
# Use custom port
./outline-hexo-connector -p 8080

# Use custom config file
./outline-hexo-connector -c /path/to/config.yaml

# Enable test mode
./outline-hexo-connector -t

# Combined usage
./outline-hexo-connector -p 8080 -c custom.yaml
```

### Configure Outline Webhook

1. Login to your Outline dashboard.
2. Go to **Settings** → **Webhooks**.
3. Create a new Webhook:
   - **URL**: `http://Outline-Hexo-Connector-IP:Port/webhook`
   - **Secret**: Copy to `Outline_Webhook_Secret` in `config.yaml`
   - **Events**: Select events to listen to. Recommended: `documents.create`, `documents.publish`, `documents.unpublish`, `documents.delete`, `documents.archive`, `documents.unarchive`, `documents.restore`, `documents.move`, `documents.update`, `documents.title_change`
4. Go to **Settings** → **API & Tokens**.
5. Create a new API Token:
   - **Scopes**: At least `documents.info`, `documents.unpublish`, `collections.info`, `attachments.redirect`
   - **Expiration**: As per your needs
   - Copy the created API Token to `Outline_API_Key` in `config.yaml`

## ⚠️ Notes

Since Outline automatically publishes newly created documents, to avoid creating a meaningless empty file in Hexo and triggering a build, this tool will automatically unpublish newly created documents. Wait until editing is complete and then publish again.

This tool will also automatically unpublish updated documents within the scope, so that users can trigger the Hexo blog build by clicking "Publish" again.

## 🏷️ Custom Document Tag Guide

To provide synced Hexo articles with complete metadata (such as tags, summary, cover image), this tool supports a set of custom Markdown syntax tags. These tags are parsed and processed during synchronization and will not be displayed directly in the article body.

### 1. Article Tags (Tags)

Used to set tags for Hexo articles. Supports separation by English comma `,` or Chinese comma `，`.

- **Syntax**: `+> Tags: tag1, tag2`
- **Position**: Recommended at the beginning or end of the document.
- **Effect**: Parsed into Front Matter as `tags: [tag1, tag2]` and removed from the body.

### 2. Summary Separator (Read More)

Controls the summary display range of the article in the home page list.

- **Syntax**: `+> More:`
- **Effect**: Replaced with Hexo's `<!-- more -->` marker. Content before this marker will be shown as the summary.

### 3. Cover & Thumbnail (Banner & Index Image)

Set the top banner image (Banner) and list thumbnail (Index Image) for the article. The syntax is similar to standard Markdown image syntax but uses specific Alt text.

| Syntax | Description |
|--------|-------------|
| `![banner_img](url)` | Sets only the article detail page top cover image (Banner) |
| `![index_img](url)` | Sets only the article list page thumbnail (Index) |
| `![banner_index_img](url)` | Sets both Banner and Index images |
| `![index_banner_img](url)` | Same as above, sets both Banner and Index images |

> **Note**: These special image tags are removed from the body after parsing and converted to Front Matter configuration.

### Example

In an Outline document:

```markdown
# My New Article

+> Tags: Technology, Golang, Tutorial

This is the summary part of the article.

+> More:

![banner_index_img](https://example.com/cover.jpg)

Here is the main content of the article...
```

## 📦 Project Structure

```
outline-hexo-connector/
├── main.go                 # Main program entry, handles args and signals
├── config_example.yaml     # Example configuration file
├── go.mod                  # Go module definition
├── README.md               # English documentation
├── README_zh.md            # Chinese documentation
└── internal/
    ├── config/
    │   └── config.go       # Configuration loading and parsing
    ├── hexo/
    │   ├── renderer.go     # Hexo post generation and writing
    │   └── trigger.go      # Hexo build triggering and debounce control
    ├── outline/
    │   ├── client.go       # Outline API client and Webhook handling
    │   └── models.go       # Outline data model definitions
    ├── processor/
    │   ├── converter.go    # Attachment URL conversion and processing
    │   └── parser.go       # Markdown content parsing and metadata extraction
    ├── server/
    │   └── direct_access.go # Direct browser access handling (HTML/proxy)
    └── test/
        └── test.go         # Testing tools and debug helpers
```

## 🛠️ Development

### Dependencies

- [pflag](https://github.com/spf13/pflag) - Command line argument parsing
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML configuration parsing

### Run Test Mode

Test mode allows you to view raw Webhook requests received:

```bash
./outline-hexo-connector -t
```

Then trigger a test event from Outline, and you will see the full request content in the console.

## 📋 Todo

- [x] Refine Hexo adapter implementation
- [x] Implement full Document to Markdown conversion
- [x] Add attachment URL conversion (from Outline API to OSS permanent link)
- [x] Trigger Hexo build on document publish/delete
- [x] Add document queue mechanism to support periodic batch builds
- [ ] Add unit tests (Maybe)
- [x] Refine error handling and logging

## 🤝 Contribution

Issues and Pull Requests are welcome!
This is a small program I wrote for practicality and to practice Go, please feel free to give advice.
~~However, basic functions haven't been fully implemented yet, will anyone really submit PRs?~~
Basic functions have been implemented, but there are still some limitations. It may be improved gradually.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Outline](https://www.getoutline.com/) - Powerful team knowledge base
- [Hexo](https://hexo.io/) - Fast, simple & powerful blog framework

## 📞 Contact

If you have questions or suggestions, please contact via:

- GitHub Issues: [https://github.com/Charles-IX/outline-hexo-connector/issues](https://github.com/Charles-IX/outline-hexo-connector/issues)

---
