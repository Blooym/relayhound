# relayhound

A simple tool to test the fanout of target event data across multiple concurrent ATProto PDS or Relay connections via the `/xrpc/com.atproto.sync.subscribeRepos` endpoint.

## Installation

```
go install codeberg.org/Blooym/relayhound@latest
```

## Usage

Relayhound has 3 flags

- `--hosts` The WebSocket URL (including protocol) to connect to (repeatable).
- `--target`: The target data (string) that you're looking for in a response.
- `--timeout`: The amount of time to keep connections open before closing them automatically (optional, default 1hour).

```
relayhound --hosts wss://<HOST_1> --hosts ws://<HOST_2> --target TARGET_DATA_STRING
```
