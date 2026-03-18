# Contributing to files2clip

Thanks for your interest in contributing!

## Development setup

1. Install [Go](https://go.dev/) 1.24+ and [Task](https://taskfile.dev/).
2. Clone the repo and run the tests:

```sh
git clone https://github.com/mmdemirbas/files2clip.git
cd files2clip
task test
```

## Running tests

```sh
task test       # run all tests
task lint       # go vet
task bench      # benchmarks
task all        # test + lint + dist
```

## Making changes

1. Create a branch from `main`.
2. Make your changes. Keep the scope focused — one feature or fix per PR.
3. Ensure `task all` passes.
4. Submit a pull request.

## Guidelines

- **No external dependencies.** The project uses stdlib only.
- **Table-driven tests** with `testing.T` and `b.Loop()` for benchmarks.
- **Internal packages** under `internal/` are not public API.
- **Errors to stderr**, normal output to stdout.
- Keep CLI output concise and machine-parseable where practical.

## Reporting issues

Open an issue with:
- What you did (command, flags, input)
- What happened vs. what you expected
- OS and Go version (`go version`)
