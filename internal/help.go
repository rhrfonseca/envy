package envy

import (
	"fmt"
	"os"
)

func StdinIsTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func Usage() {
	fmt.Print(`envy — substitute ${ENV_VARS} in a YAML file

Usage:
  envy [flags] <file.yaml>
  cat file.yaml | envy [flags]

Flags:
  --env <name>        environment name; loads .env.<name> and .env.<name>.local
  --dotenv <file>     explicit dotenv file (repeatable; first flag wins; replaces standard lookup)
  --tmp               write output to /tmp/envy-<name>-<timestamp>.yaml and print the path
  --out, -o <file>    write output to a file (silent)

Variable syntax:
  ${VAR}              substituted with the value of VAR
  ${VAR:-default}     uses "default" if VAR is not set
  $${VAR}             escaped; outputs literal ${VAR}

Dotenv precedence (highest to lowest):
  1. Shell environment
  2. .env.<env>.local  (only when --env is set)
  3. .env.local
  4. .env.<env>        (only when --env is set)
  5. .env

  Files are looked up in the current working directory.
  Passing --dotenv replaces the standard lookup entirely.

Examples:
  envy config.yaml
  envy --env production config.yaml
  envy --dotenv secrets.env --dotenv base.env config.yaml
  envy --tmp config.yaml | xargs kubectl apply -f
  cat config.yaml | envy > out.yaml
`)
}
