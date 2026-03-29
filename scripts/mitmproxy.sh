#!/usr/bin/env bash
set -euo pipefail

mode="${1:-web}"

common_args=(
  --mode reverse:https://api.anthropic.com@3000
  --set stream_large_bodies=1m
  --set store_streamed_bodies=true
)

case "$mode" in
  web)
    uvx --from mitmproxy mitmweb \
      "${common_args[@]}" \
      --listen-host 0.0.0.0 \
      --set web_host=0.0.0.0 \
      --set web_password=password \
      --set web_allow_hosts='.*'
    ;;
  proxy)
    uvx --from mitmproxy mitmproxy \
      "${common_args[@]}"
    ;;
  *)
    echo "Usage: $0 [web|proxy]" >&2
    exit 1
    ;;
esac
