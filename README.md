# relayhound

A simple tool to test the fanout of target event data across multiple concurrent ATProto PDS or Relay connections via the `/xrpc/com.atproto.sync.subscribeRepos` endpoint.

## Installation

```
go install codeberg.org/Blooym/relayhound@latest
```

## Usage

Relayhound has 2 flags

- `--hosts` to specify the URL to one or more relays (including Protocol). This flag is repeatable.
- `--target` to specify the target data (string) that you're looking for in a response.

```
relayhound --hosts wss://<HOST_1> --hosts ws://<HOST_2> --target TARGET_DATA_STRING
```
