# nvim-simple-keybind-helper

Search over your keybinds if you forget them.

## Install

1. Install:

```bash
go install github.com/JNickson/nvim-simple-keybind-helper@latest
```

2. Run it from your shell (or add an alias):

```bash
nvim-simple-keybind-helper
```

## Development

```bash
make fmt
make test
make build
```

## Config from JSON

By default, the app uses the built-in keybind table from code.

- To use your own config file once: `nvim-simple-keybind-helper --config /your/path/config.json`
- To make it persistent, set `NVIM_HELPER_CONFIG=/your/path/config.json`
- Precedence: `--config` flag overrides `NVIM_HELPER_CONFIG`.
- If neither is set, no file is read.

Alias example:

```bash
alias vim-keybinds='nvim-simple-keybind-helper --config "$HOME/.config/nvim-helper/my-bindings.json"'
```

Config shape:

```json
{
  "columns": [
    { "title": "Mode", "width": 8 },
    { "title": "Keybind", "width": 16 },
    { "title": "Action", "width": 80 }
  ],
  "rows": [
    { "mode": "normal", "keybind": "gd", "action": "go to definition" }
  ],
  "height": 7
}
```
