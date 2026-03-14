# AGENTS.md

This file is the working agreement for changes in this repository.

## Workflow

Default workflow for this repo:

1. Make the requested code or docs change directly.
2. Verify it thoroughly.
3. Commit the change before finishing.

Do not stop at unit tests when behavior can be exercised for real. If a change affects runtime behavior, flags, mounts, image builds, config loading, or release automation, run the actual path and verify the output.

If full end-to-end verification is blocked by the environment, state exactly what was run, what was not run, and why.

Any hard-earned lesson that changes how future work should be done belongs in `AGENTS.md`.

## Build Commands

```bash
make build          # Build the yolobox binary
make test           # Run unit tests
make lint           # Run go vet (and golangci-lint if installed)
make image          # Build the Docker base image
make smoke-test     # Run smoke tests on container tools
make install        # Build and install to ~/.local/bin
make clean          # Remove built binary
```

## Verification Standard

For non-trivial code changes, start here:

```bash
make clean && make build && make test
./yolobox version
./yolobox help
./yolobox config
```

Then run the feature you changed end-to-end. Examples:

- Runtime or mount behavior: `./yolobox run ...`
- Docker integration: `./yolobox run --docker docker version`
- Custom image flow: `./yolobox run --packages cowsay /usr/games/cowsay hello`
- Config loading: `./yolobox config`

For Dockerfile or image changes, rebuild the image and run container smoke tests:

```bash
make image
./yolobox run echo hello
./yolobox run whoami
./yolobox run pwd
```

For docs-only changes, inspect the diff carefully. Tests are optional unless the docs describe behavior that also changed.

## Code Map

Most logic lives in [cmd/yolobox/main.go](cmd/yolobox/main.go).

Areas to check when adding flags or config:

- `Config`
- `parseBaseFlags`
- `mergeConfig`
- `printConfig`
- `saveGlobalConfig`
- `runSetup`
- the runtime path that consumes the option

Also update [README.md](README.md) when user-facing behavior changes.

## Hard Learnings

- Named volumes shadow image contents. Anything baked into `/home/yolo` disappears behind `yolobox-home` for existing users.
- On macOS, Docker socket source paths are resolved inside the Docker VM. Use `/var/run/docker.sock` as the mount source, not host-side paths like `~/.colima/default/docker.sock`.
- On macOS, `SSH_AUTH_SOCK` from the host is not directly mountable into Docker. Docker Desktop uses `/run/host-services/ssh-auth.sock`; Colima requires `forwardAgent: true` and a VM-side socket path from `colima ssh -- printenv SSH_AUTH_SOCK`.
- Claude config is split across `~/.claude/` and `~/.claude.json`, and the config directory must be writable inside the container.
- Claude OAuth creds on macOS live in Keychain, not just on disk.
- `gh` tokens on macOS may also live in Keychain; `gh auth token` is the reliable extraction path.
- Colima often defaults to 2GB RAM, which is too small for heavier agent workflows. 4GB+ is the practical floor.
- Global npm installs as `yolo` need a user-writable prefix such as `/home/yolo/.npm-global`.
- If a Docker build gets SIGKILL while equivalent runtime commands succeed, split heavy installers into a separate stage to reduce layer memory pressure.
