# tailer
A tool for watch log files from multiple servers.

# Build

1. `go mod tidy`
2. `go build`

# Usage

`tailer` for help

Examples

`tailer tailf test /data/nginx/access.log`

`tailer tailf test nginx_access_log`

`tailer -config /etc/tailer.yaml tailf another_project another_log_file`

# Config

Default is `~/.tailer.yaml`(if exists). You can special a config file with `-config` flag

