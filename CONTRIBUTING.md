# Contributing

## Adding a code

Every code lives in `codes.json`, which is embedded into the binary at
build time. You do not need to touch any Go code.

1. Find the right table. `tables` holds the codes that apply everywhere
   (shell, sysexits). `contexts` holds codes for one tool, selected with
   `--context`. `signals` holds the signals behind 128+N codes, with their
   Linux and macOS numbers.
2. Add an entry, keeping the table sorted by code:

   ```json
   {"code": 7, "name": "CURLE_COULDNT_CONNECT", "meaning": "failed to connect to the host", "hints": ["Check that the service is listening on that port."]}
   ```

   - `meaning` is a short lowercase phrase that reads well after
     `7 (CURLE_COULDNT_CONNECT): `. No trailing period.
   - `hints` holds one or two concrete next steps, as full sentences.
     Put commands in backticks.
   - `name` is optional, for codes that have a symbolic name.
   - Set `"signal": true` when the code means 128+N, so the signal name
     is shown next to it.
3. Run `go test ./...`. The tests check every entry for range,
   duplicates, empty text and hint count.

## Adding a context

Add an object to `contexts` with an `id` (the `--context` value), a
`title`, the `commands` the shell hook should match, and its `codes`. Set
`"os": "linux"` if the tool always reports Linux signal numbers, as
containers do. Then add a row to the contexts table in the README.

## Other changes

Run `gofmt`, `go vet ./...` and `go test ./...` before sending a pull
request. Use conventional commit messages, such as `feat: add rsync
codes` or `fix: keep $? in fish hook`.
