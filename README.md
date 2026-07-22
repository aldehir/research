# Research Paper Reader

A self-hosted web app for reading research papers with Claude at your side.
Upload PDFs, read them in the browser, select a passage, and ask questions
about what you're reading. Everything runs from a single server binary on your
own machine, with no accounts and no cloud storage.

## Features

- **Personal library**: upload papers by drag-and-drop and access them from
  any device on your network.
- **In-browser reading**: full PDF rendering with text selection, zoom, and
  page navigation.
- **Chat with Claude**: ask questions about the paper; selected passages and
  surrounding context are sent along for better answers.
- **Local storage**: papers live on your filesystem and metadata in a SQLite
  file, so backup is just copying files.

## Requirements

To build the app you need:

- [Go](https://go.dev/) 1.25 or newer
- [Node.js](https://nodejs.org/) with [pnpm](https://pnpm.io/)

To run the server you need these command-line tools installed:

- `pdftotext` and `pdftoppm`, part of [Poppler](https://poppler.freedesktop.org/)
  (`poppler-utils` on most Linux distributions, `poppler` on Homebrew)
- [`qpdf`](https://qpdf.sourceforge.io/)

To use the chat features you also need an
[Anthropic API key](https://console.anthropic.com/).

## Building

Build the frontend first, then the server. The server embeds the frontend, so
the result is a single binary:

```sh
cd frontend
pnpm install
pnpm build
cd ..
go build ./cmd/research-server
```

This produces a `research-server` binary in the project root.

## Running

Set your API key and start the server:

```sh
export ANTHROPIC_API_KEY=sk-ant-...
./research-server
```

Then open <http://localhost:8080> in your browser. On first run the server
creates a `research.db` database file and a `./data` directory for your PDFs
in the current directory.

Without an API key the app still works as a PDF reader; only chat is
disabled.

## Configuration

All settings are optional and can be given as flags or environment variables:

| Flag             | Environment variable | Default       | Purpose                              |
|------------------|----------------------|---------------|--------------------------------------|
| `--addr`         | `ADDR`               | `:8080`       | Address and port to listen on        |
| `--db-path`      | `DB_PATH`            | `research.db` | Where the database file is stored    |
| `--data-dir`     | `DATA_DIR`           | `./data`      | Where PDFs and attachments are kept  |
| `--log-level`    | `LOG_LEVEL`          | `info`        | Log detail: debug, info, warn, error |

Chat behavior is controlled with environment variables:

| Variable                | Purpose                                             |
|-------------------------|-----------------------------------------------------|
| `ANTHROPIC_API_KEY`     | Enables chat with Claude                            |
| `ANTHROPIC_MODEL`       | Use a specific Claude model instead of the default  |
| `RESEARCH_INSTRUCTIONS` | Extra instructions added to every chat conversation |

Run `./research-server --help` for the full list of flags.

## License

MIT. See [LICENSE](LICENSE).
