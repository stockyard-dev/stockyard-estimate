# Stockyard Estimate

**Self-hosted estimates and quotes with line items**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted tools.

## Quick Start

```bash
curl -fsSL https://stockyard.dev/tools/estimate/install.sh | sh
```

Or with Docker:

```bash
docker run -p 9802:9802 -v estimate_data:/data ghcr.io/stockyard-dev/stockyard-estimate
```

Open `http://localhost:9802` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9802` | HTTP port |
| `DATA_DIR` | `./estimate-data` | SQLite database directory |
| `STOCKYARD_LICENSE_KEY` | *(empty)* | License key for unlimited use |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 5 records | Unlimited |
| Price | Free | Included in bundle or $29.99/mo individual |

Get a license at [stockyard.dev](https://stockyard.dev).

## License

Apache 2.0
