# whyexit

`whyexit` tells you what a process exit code means and what to check next.

```console
$ whyexit 137
137 = 128 + 9 (SIGKILL): killed by force, often by the OOM killer
  Check `dmesg | grep -i oom` or `journalctl -k` for an out-of-memory kill.
  Otherwise something ran `kill -9`, such as a timeout or a supervisor.
```

## Install

Homebrew:

```sh
brew install fuatbeser/tap/whyexit
```

Install script (Linux, macOS, FreeBSD; checks the release checksum):

```sh
curl -fsSL https://raw.githubusercontent.com/fuatbeser/whyexit/main/install.sh | sh
```

With Go 1.22 or later:

```sh
go install github.com/fuatbeser/whyexit@latest
```

Prebuilt binaries for every platform, Windows included, are on the
[releases page](https://github.com/fuatbeser/whyexit/releases).

![whyexit demo](demo.gif)

## Usage

```console
$ whyexit 127
127: command not found
  Check the spelling and that the directory is on $PATH.
  In scripts, a missing dependency or a wrong shebang path often causes this.

$ whyexit --context docker 137
docker: 137 = 128 + 9 (SIGKILL): the container was killed, often for running out of memory
  Run `docker inspect <container> --format '{{.State.OOMKilled}}'`; true means it hit its memory limit.
  If false, `docker stop` timed out waiting or something ran `docker kill`.

$ whyexit --context curl 6
curl: 6 (CURLE_COULDNT_RESOLVE_HOST): the host could not be resolved
  Check the host name for typos and that DNS works: `host <name>`.
  In containers, check /etc/resolv.conf and the network mode.
```

Without `--context`, codes are looked up in the shell table, then BSD
`sysexits.h` (64-78), then as 128+N for a signal. Signal numbers differ
between Linux and macOS (138 is SIGUSR1 on Linux and SIGBUS on macOS), so
whyexit uses the numbering of the system it runs on. Pass `--os linux` or
`--os macos` to override it. The docker and k8s contexts always use Linux
numbering, since containers run Linux.

Other flags:

- `--json` prints the explanation as JSON, for scripts and agents.
- `--list` prints every known code; combine it with `--context` or `--json`.
- Codes outside 0-255 are wrapped the way the shell sees them, so 256 is 0.

## Contexts

| Context  | Covers                                               | Hook matches                    |
|----------|------------------------------------------------------|---------------------------------|
| `docker` | 125, 126, 127 from the CLI; 1, 137, 139, 143 from containers | `docker`, `podman`              |
| `k8s`    | container exit codes in pod status: 1, 126, 127, 134, 137, 139, 143 | `kubectl`                       |
| `curl`   | every curl exit code from 1 to 99                    | `curl`                          |
| `git`    | 1, 128 (fatal), 129 (usage)                          | `git`                           |
| `npm`    | npm script failures and node's own exit codes        | `npm`, `npx`, `node`, `yarn`, `pnpm` |
| `ssh`    | 255 from ssh itself, and remote status pass-through  | `ssh`                           |

## Shell hook

The hook prints one dim line after any command that fails:

```console
$ curl -s localhost:1
curl exit 7 (CURLE_COULDNT_CONNECT): failed to connect to the host
```

Add one of these to your shell's startup file:

```sh
eval "$(whyexit init zsh)"     # ~/.zshrc
eval "$(whyexit init bash)"    # ~/.bashrc
whyexit init fish | source     # ~/.config/fish/config.fish
```

The hook picks the context from the command that failed, keeps `$?` (or
`$status`) intact for your prompt, and only runs whyexit when a command
fails, so a successful command costs nothing. It stays quiet for Ctrl+C
and for jobs stopped with Ctrl+Z. Use `whyexit init --show-interrupt zsh`
to get a hint for Ctrl+C as well. In bash the hook reads the failed
command from history, so with history turned off it still explains the
code but cannot pick a context.

Set `NO_COLOR` to turn off the dim styling.

## Contributing

All codes live in [codes.json](codes.json); adding one needs no Go
changes. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
