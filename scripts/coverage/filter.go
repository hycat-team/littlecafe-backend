// Command filter trims a Go coverage profile to the project's business-code
// scope and reports the resulting statement coverage.
//
// It is run directly with `go run scripts/coverage/filter.go` so it stays
// outside every workspace module.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultExcludes = "cmd/,internal/di/,internal/bootstrap/,database/migrations/,shared/infrastructure/,contracts/openapi/,shared/provider.go,_gen.go,.pb.go"

func main() {
	in := flag.String("in", "coverage/coverage.out", "input coverage profile")
	out := flag.String("out", "coverage/coverage.business.out", "output coverage profile")
	exclude := flag.String("exclude", defaultExcludes, "comma-separated path fragments to exclude")
	min := flag.Float64("min", 0, "minimum total statement coverage percentage; 0 disables the check")
	flag.Parse()

	excludes := make([]string, 0, 16)
	for _, item := range strings.Split(*exclude, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			excludes = append(excludes, item)
		}
	}

	src, err := os.Open(*in)
	if err != nil {
		fatal("open profile: %v", err)
	}
	defer src.Close()

	var (
		output    strings.Builder
		totalStmt int64
		coverStmt int64
		seenMode  bool
	)
	scanner := bufio.NewScanner(src)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			output.WriteString(line)
			output.WriteByte('\n')
			seenMode = true
			continue
		}
		if line == "" {
			continue
		}
		if isExcluded(line, excludes) {
			continue
		}
		stmt, count, ok := parseBlock(line)
		if !ok {
			continue
		}
		totalStmt += stmt
		if count > 0 {
			coverStmt += stmt
		}
		output.WriteString(line)
		output.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		fatal("read profile: %v", err)
	}
	if !seenMode {
		fatal("profile %s has no mode header", *in)
	}
	if totalStmt == 0 {
		fatal("profile %s has no statements inside the coverage scope", *in)
	}
	if err := os.WriteFile(*out, []byte(output.String()), 0o644); err != nil {
		fatal("write filtered profile: %v", err)
	}

	percent := float64(coverStmt) / float64(totalStmt) * 100
	fmt.Printf("total:\t\t\t\t\t\t\t\t\t\t(statements)\t\t%.1f%%\n", percent)
	fmt.Printf("scope: %d/%d statements kept in %s\n", coverStmt, totalStmt, *out)

	if *min > 0 && percent < *min {
		fmt.Fprintf(os.Stderr, "coverage %.1f%% is below the required %.1f%%\n", percent, *min)
		os.Exit(1)
	}
}

func isExcluded(line string, excludes []string) bool {
	for _, fragment := range excludes {
		if strings.Contains(line, fragment) {
			return true
		}
	}
	return false
}

func parseBlock(line string) (int64, int64, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return 0, 0, false
	}
	stmt, err := strconv.ParseInt(fields[len(fields)-2], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	count, err := strconv.ParseInt(fields[len(fields)-1], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return stmt, count, true
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
