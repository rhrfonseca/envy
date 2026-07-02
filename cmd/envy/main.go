package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	envy "github.com/rhrfonseca/envy/internal"
	"gopkg.in/yaml.v3"
)

type stringSliceFlag []string

func (f *stringSliceFlag) String() string {
	return strings.Join(*f, ", ")
}

func (f *stringSliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	var (
		envName     string
		dotenvFiles stringSliceFlag
		tmp         bool
		out         string
	)

	flag.Usage = envy.Usage
	flag.StringVar(&envName, "env", "", "environment name (e.g. production)")
	flag.Var(&dotenvFiles, "dotenv", "dotenv file path (repeatable, first flag wins, replaces standard lookup)")
	flag.BoolVar(&tmp, "tmp", false, "write output to /tmp/envy-<name>-<timestamp>.yaml and print the path")
	flag.StringVar(&out, "out", "", "write output to specified file path (silent)")
	flag.StringVar(&out, "o", "", "write output to specified file path (silent)")

	flag.Parse()

	if flag.NArg() == 0 && envy.StdinIsTerminal() {
		envy.Usage()
		os.Exit(0)
	}

	if tmp && out != "" {
		fmt.Fprintln(os.Stderr, "error: --tmp and --out/-o are mutually exclusive")
		os.Exit(1)
	}

	var input []byte
	var inputName string

	args := flag.Args()
	if len(args) > 0 {
		inputFile := args[0]
		base := filepath.Base(inputFile)
		inputName = strings.TrimSuffix(base, filepath.Ext(base))
		var err error
		input, err = os.ReadFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	} else {
		inputName = "stdin"
		var err error
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
			os.Exit(1)
		}
	}

	var beforeCheck interface{}
	if err := yaml.Unmarshal(input, &beforeCheck); err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid YAML input: %v\n", err)
		os.Exit(1)
	}

	envMap, err := envy.LoadEnvMap(envName, []string(dotenvFiles))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	result, missingVars := envy.Substitute(string(input), envMap)

	if len(missingVars) > 0 {
		sort.Strings(missingVars)
		fmt.Fprintln(os.Stderr, "error: the following variables are not defined:")
		for _, v := range missingVars {
			fmt.Fprintf(os.Stderr, "  - %s\n", v)
		}
		os.Exit(1)
	}

	var afterCheck interface{}
	if err := yaml.Unmarshal([]byte(result), &afterCheck); err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid YAML after substitution: %v\n", err)
		os.Exit(1)
	}

	if tmp {
		timestamp := time.Now().Unix()
		tmpPath := fmt.Sprintf("/tmp/envy-%s-%d.yaml", inputName, timestamp)
		if err := os.WriteFile(tmpPath, []byte(result), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing tmp file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(tmpPath)
	} else if out != "" {
		if err := os.WriteFile(out, []byte(result), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing output file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(result)
	}
}
