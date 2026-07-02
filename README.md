# envy

Substitute `${ENV_VARS}` in a YAML file using values from your environment or dotenv files.

## Installation

### From source

Requires [Go 1.21+](https://go.dev/dl/).

```sh
git clone https://github.com/rhrfonseca/envy.git
cd envy
make install        # installs to /usr/local/bin
# or
PREFIX=~/bin make install
```

### With go install

```sh
go install github.com/rhrfonseca/envy/cmd/envy@latest
```

## Usage

```sh
envy [flags] <file.yaml>
cat file.yaml | envy [flags]
```

### Flags

| Flag | Description |
|---|---|
| `--env <name>` | Environment name — loads `.env.<name>` and `.env.<name>.local` |
| `--dotenv <file>` | Explicit dotenv file (repeatable; first flag wins; replaces standard lookup) |
| `--tmp` | Write output to `/tmp/envy-<name>-<timestamp>.yaml` and print the path |
| `--out`, `-o <file>` | Write output to a file (silent) |

### Variable syntax

| Syntax | Behaviour |
|---|---|
| `${VAR}` | Replaced with the value of `VAR` |
| `${VAR:-default}` | Uses `default` when `VAR` is not set |
| `$${VAR}` | Escaped — outputs the literal string `${VAR}` |

Substitution happens anywhere in the file: values, keys, and comments. All missing variables are reported at once before exiting.

The YAML is validated before and after substitution. A substitution that produces invalid YAML is an error.

### Dotenv precedence

When no `--dotenv` flag is given, `envy` looks for these files in the current working directory (highest priority first):

1. Shell environment
2. `.env.<env>.local` *(only when `--env` is set)*
3. `.env.local`
4. `.env.<env>` *(only when `--env` is set)*
5. `.env`

Passing `--dotenv` replaces the standard lookup entirely. Multiple `--dotenv` files are supported; the first flag takes precedence.

## Examples

```sh
# Basic substitution
envy config.yaml

# Production environment
envy --env production config.yaml

# Multiple explicit dotenv files (secrets.env wins over base.env)
envy --dotenv secrets.env --dotenv base.env config.yaml

# Write to a temp file (useful when another command expects a path)
kubectl apply -f $(envy --tmp config.yaml)

# Pipe
cat config.yaml | envy > out.yaml

# Write to an explicit output file
envy --out rendered.yaml config.yaml
```

## Contributing

1. Fork the repository and create a branch from `main`.
2. Make your changes, add tests, and ensure the suite passes:
   ```sh
   make test
   ```
3. Open a pull request with a clear description of what changed and why.

Please add a `LICENSE` file before publishing if one is not already present.
