package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	sqlformatter "sql-formatter-go"
)

var version = "dev"

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	output := fs.String("output", "", "File to write SQL output (defaults to stdout)")
	outputShort := fs.String("o", "", "File to write SQL output (defaults to stdout)")
	fix := fs.Bool("fix", false, "Update the file in-place")
	check := fs.Bool("check", false, "Check if files are formatted (exit 1 if not)")
	lang := fs.String("language", "sql", "SQL dialect (defaults to basic sql)")
	langShort := fs.String("l", "", "SQL dialect (defaults to basic sql)")
	config := fs.String("config", "", "Path to config JSON file or json string (will find a file named '.sql-formatter.json' or use default configs if unspecified)")
	configShort := fs.String("c", "", "Path to config JSON file or json string (will find a file named '.sql-formatter.json' or use default configs if unspecified)")
	showVersion := fs.Bool("version", false, "show program's version number and exit")
	workers := fs.Int("workers", 0, "number of concurrent workers for multiple files (0 = NumCPU)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: %s [-h] [-o OUTPUT] [-l {postgresql,sql}] [-c CONFIG] [--version] [FILE...]\n\n", os.Args[0])
		fmt.Fprintln(fs.Output(), "SQL Formatter")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "positional arguments:")
		fmt.Fprintln(fs.Output(), "  FILE            Input SQL file(s) (defaults to stdin)")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "optional arguments:")
		fmt.Fprintln(fs.Output(), "  -h, --help      show this help message and exit")
		fmt.Fprintln(fs.Output(), "  -o, --output    OUTPUT")
		fmt.Fprintln(fs.Output(), "                    File to write SQL output (defaults to stdout)")
		fmt.Fprintln(fs.Output(), "  --fix           Update the file in-place")
		fmt.Fprintln(fs.Output(), "  --check         Check if files are formatted (exit 1 if not)")
		fmt.Fprintln(fs.Output(), "  -l, --language  {postgresql,sql}")
		fmt.Fprintln(fs.Output(), "                    SQL dialect (defaults to basic sql)")
		fmt.Fprintln(fs.Output(), "  -c, --config    CONFIG")
		fmt.Fprintln(fs.Output(), "                    Path to config JSON file or json string (will find a file named '.sql-formatter.json' or use default configs if unspecified)")
		fmt.Fprintln(fs.Output(), "  --workers       number of concurrent workers for multiple files (0 = NumCPU)")
		fmt.Fprintln(fs.Output(), "  --version       show program's version number and exit")
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}

	if *showVersion {
		fmt.Println(version)
		return
	}

	files := fs.Args()

	if *output == "" && *outputShort != "" {
		*output = *outputShort
	}
	if *config == "" && *configShort != "" {
		*config = *configShort
	}
	if *lang == "sql" && *langShort != "" {
		*lang = *langShort
	}

	if *output != "" && *fix {
		fmt.Fprintln(os.Stderr, "Error: Cannot use both --output and --fix options simultaneously")
		os.Exit(1)
	}
	if *check && *fix {
		fmt.Fprintln(os.Stderr, "Error: Cannot use both --check and --fix options simultaneously")
		os.Exit(1)
	}
	if *check && *output != "" {
		fmt.Fprintln(os.Stderr, "Error: Cannot use both --check and --output options simultaneously")
		os.Exit(1)
	}
	if *fix && len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: The --fix option cannot be used without a filename")
		os.Exit(1)
	}
	if *check && len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: The --check option cannot be used without a filename")
		os.Exit(1)
	}
	if *output != "" && len(files) > 1 {
		fmt.Fprintln(os.Stderr, "Error: The --output option cannot be used with multiple input files")
		os.Exit(1)
	}

	if isTTY(os.Stdin) && len(files) == 0 && *output == "" && !*fix && *config == "" {
		fs.Usage()
		return
	}

	cfgMap, err := loadConfig(*config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cfg, err := buildConfig(*lang, cfgMap)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if len(files) == 0 {
		query, err := readInput("")
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		formatted, err := sqlformatter.Format(query, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		formatted = strings.TrimSpace(formatted) + "\n"
		if *output == "" {
			_, _ = os.Stdout.WriteString(formatted)
			return
		}
		if err := os.WriteFile(*output, []byte(formatted), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not write file %s\n", *output)
			os.Exit(1)
		}
		return
	}

	runMultiFile(files, cfg, *fix, *check, *output, *workers)
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func readInput(file string) (string, error) {
	if file == "" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("Error: unable to read stdin")
		}
		if len(data) == 0 && isTTY(os.Stdin) {
			return "", errors.New("Error: no file specified and no data in stdin")
		}
		return string(data), nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("Error: could not open file %s", file)
	}
	return string(data), nil
}

func runMultiFile(files []string, cfg sqlformatter.FormatOptionsWithLanguage, fix bool, check bool, output string, workers int) {
	workerCount := workers
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
		if workerCount < 1 {
			workerCount = 1
		}
	}
	type result struct {
		index     int
		text      string
		err       error
		file      string
		different bool // for check mode: true if file differs from formatted
	}

	jobs := make(chan int, len(files))
	results := make(chan result, len(files))
	var failed int64

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for idx := range jobs {
				path := files[idx]
				data, err := os.ReadFile(path)
				if err != nil {
					results <- result{index: idx, err: fmt.Errorf("Error: could not open file %s", path), file: path}
					continue
				}
				formatted, err := sqlformatter.Format(string(data), cfg)
				if err != nil {
					results <- result{index: idx, err: err, file: path}
					continue
				}
				formatted = strings.TrimSpace(formatted) + "\n"
				if check {
					different := string(data) != formatted
					results <- result{index: idx, file: path, different: different}
					continue
				}
				if fix {
					if err := os.WriteFile(path, []byte(formatted), 0o644); err != nil {
						results <- result{index: idx, err: fmt.Errorf("Error: could not write file %s", path), file: path}
						continue
					}
					results <- result{index: idx, text: "", file: path}
					continue
				}
				results <- result{index: idx, text: formatted, file: path}
			}
		}()
	}

	for i := range files {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	close(results)

	out := make([]string, len(files))
	var checkFailed int64
	for res := range results {
		if res.err != nil {
			atomic.AddInt64(&failed, 1)
			fmt.Fprintln(os.Stderr, res.err.Error())
			continue
		}
		if check && res.different {
			atomic.AddInt64(&checkFailed, 1)
			fmt.Fprintf(os.Stderr, "%s\n", res.file)
		}
		out[res.index] = res.text
	}

	if failed > 0 {
		os.Exit(1)
	}

	if check {
		if checkFailed > 0 {
			os.Exit(1)
		}
		return
	}

	if fix {
		return
	}

	if output != "" {
		if err := os.WriteFile(output, []byte(out[0]), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not write file %s\n", output)
			os.Exit(1)
		}
		return
	}

	for _, text := range out {
		if text == "" {
			continue
		}
		_, _ = os.Stdout.WriteString(text)
	}
}

func loadConfig(configArg string) (map[string]interface{}, error) {
	if configArg != "" {
		var cfg map[string]interface{}
		if err := json.Unmarshal([]byte(configArg), &cfg); err == nil {
			return cfg, nil
		}
		data, err := os.ReadFile(configArg)
		if err == nil {
			if err := json.Unmarshal(data, &cfg); err == nil {
				return cfg, nil
			}
		}
		return nil, fmt.Errorf("Error: unable to parse as JSON or treat as JSON file: %s", configArg)
	}

	cfgPath := findConfigFile()
	if cfgPath == "" {
		return nil, nil
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("Error: unable to parse as JSON or treat as JSON file: %s", cfgPath)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("Error: unable to parse as JSON or treat as JSON file: %s", cfgPath)
	}
	return cfg, nil
}

func findConfigFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		path := filepath.Join(dir, ".sql-formatter.json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func buildConfig(cliLang string, cfg map[string]interface{}) (sqlformatter.FormatOptionsWithLanguage, error) {
	lang := cliLang
	if cfgLang, ok := cfg["language"].(string); ok && cfgLang != "" {
		lang = cfgLang
	}
	switch strings.ToLower(lang) {
	case "postgresql", "sql":
		// both map to postgres for now
	default:
		return sqlformatter.FormatOptionsWithLanguage{}, fmt.Errorf("Unsupported SQL dialect: %s", lang)
	}

	opts, err := parseFormatOptions(cfg)
	if err != nil {
		return sqlformatter.FormatOptionsWithLanguage{}, err
	}
	return sqlformatter.FormatOptionsWithLanguage{
		Language:      sqlformatter.SqlLanguage(strings.ToLower(lang)),
		FormatOptions: opts,
	}, nil
}

func parseFormatOptions(cfg map[string]interface{}) (sqlformatter.FormatOptions, error) {
	var opts sqlformatter.FormatOptions
	if cfg == nil {
		return opts, nil
	}

	if v, ok := cfg["tabWidth"].(float64); ok {
		opts.TabWidth = int(v)
	}
	if v, ok := cfg["useTabs"].(bool); ok {
		opts.UseTabs = v
	}
	if v, ok := cfg["keywordCase"].(string); ok {
		opts.KeywordCase = sqlformatter.KeywordCase(strings.ToLower(v))
	}
	if v, ok := cfg["identifierCase"].(string); ok {
		opts.IdentifierCase = sqlformatter.IdentifierCase(strings.ToLower(v))
	}
	if v, ok := cfg["dataTypeCase"].(string); ok {
		opts.DataTypeCase = sqlformatter.DataTypeCase(strings.ToLower(v))
	}
	if v, ok := cfg["functionCase"].(string); ok {
		opts.FunctionCase = sqlformatter.FunctionCase(strings.ToLower(v))
	}
	if v, ok := cfg["indentStyle"].(string); ok {
		opts.IndentStyle = sqlformatter.IndentStyle(v)
	}
	if v, ok := cfg["logicalOperatorNewline"].(string); ok {
		opts.LogicalOperatorNewline = sqlformatter.LogicalOperatorNewline(strings.ToLower(v))
	}
	if v, ok := cfg["expressionWidth"].(float64); ok {
		opts.ExpressionWidth = int(v)
		opts.ExpressionWidthSet = true
	}
	if v, ok := cfg["linesBetweenQueries"].(float64); ok {
		opts.LinesBetweenQueries = int(v)
		opts.LinesBetweenQueriesSet = true
	}
	if v, ok := cfg["denseOperators"].(bool); ok {
		opts.DenseOperators = v
	}
	if v, ok := cfg["newlineBeforeSemicolon"].(bool); ok {
		opts.NewlineBeforeSemicolon = v
	}
	if v, ok := cfg["params"]; ok {
		switch val := v.(type) {
		case []interface{}:
			list := make([]string, 0, len(val))
			for _, item := range val {
				s, ok := item.(string)
				if !ok {
					list = nil
					break
				}
				list = append(list, s)
			}
			if list != nil {
				opts.Params = list
			}
		case map[string]interface{}:
			m := make(map[string]string)
			allStrings := true
			for k, raw := range val {
				s, ok := raw.(string)
				if !ok {
					allStrings = false
					break
				}
				m[k] = s
			}
			if allStrings {
				opts.Params = m
			} else {
				opts.Params = val
			}
		}
	}
	if v, ok := cfg["paramTypes"].(map[string]interface{}); ok {
		opts.ParamTypes = parseParamTypes(v)
	}
	return opts, nil
}

func parseParamTypes(raw map[string]interface{}) *sqlformatter.ParamTypes {
	if raw == nil {
		return nil
	}
	pt := &sqlformatter.ParamTypes{}
	if v, ok := raw["positional"].(bool); ok && v {
		pt.Positional = true
	}
	if v, ok := raw["numbered"].([]interface{}); ok {
		pt.Numbered = toStringSlice(v)
	}
	if v, ok := raw["named"].([]interface{}); ok {
		pt.Named = toStringSlice(v)
	}
	if v, ok := raw["quoted"].([]interface{}); ok {
		pt.Quoted = toStringSlice(v)
	}
	if v, ok := raw["custom"].([]interface{}); ok {
		custom := make([]sqlformatter.CustomParameter, 0, len(v))
		for _, item := range v {
			obj, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			regex, _ := obj["regex"].(string)
			if regex == "" {
				continue
			}
			custom = append(custom, sqlformatter.CustomParameter{Regex: regex})
		}
		if len(custom) > 0 {
			pt.Custom = custom
		}
	}
	if !pt.Positional && len(pt.Numbered) == 0 && len(pt.Named) == 0 && len(pt.Quoted) == 0 && len(pt.Custom) == 0 {
		return nil
	}
	return pt
}

func toStringSlice(raw []interface{}) []string {
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
