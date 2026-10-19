package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/lone4alker/bootwhy/internal/acquire"
	"github.com/lone4alker/bootwhy/internal/analyze"
	"github.com/lone4alker/bootwhy/internal/diagnose"
	"github.com/lone4alker/bootwhy/internal/parse"
	"github.com/lone4alker/bootwhy/internal/report"
	"github.com/lone4alker/bootwhy/internal/rules"
)

var version = "dev"

func main() {
	args := os.Args[1:]
	command := "report"
	if len(args) > 0 && args[0] == "explain" {
		command, args = "explain", args[1:]
	}
	flags := flag.NewFlagSet("bootwhy", flag.ExitOnError)
	from := flags.String("from", "", "read a captured boot from DIR")
	showVersion := flags.Bool("version", false, "print version")
	jsonOutput := flags.Bool("json", false, "write JSON")
	flags.Parse(args)

	if *showVersion {
		fmt.Println(version)
		return
	}
	if *from == "" {
		if _, err := exec.LookPath("systemd-analyze"); err != nil {
			fmt.Fprintln(os.Stderr, "BootWhy needs systemd. This system does not have systemd-analyze.")
			os.Exit(2)
		}
	}

	raw, err := acquire.Load(context.Background(), *from)
	if err != nil {
		fmt.Fprintln(os.Stderr, "BootWhy:", err)
		os.Exit(2)
	}
	times, err := parse.ParseTime(bytes.NewReader(raw.Time))
	if err != nil {
		if errors.Is(err, parse.ErrBootInProgress) {
			fmt.Fprintln(os.Stderr, "BootWhy: boot is still in progress; wait a minute and try again.")
		} else {
			fmt.Fprintln(os.Stderr, "BootWhy: could not read boot timing data:", err)
		}
		os.Exit(2)
	}
	chain, err := parse.ParseChain(bytes.NewReader(raw.Chain))
	if err != nil {
		fmt.Fprintln(os.Stderr, "BootWhy: could not read the critical chain:", err)
		os.Exit(2)
	}
	blame, err := parse.ParseBlame(bytes.NewReader(raw.Blame))
	if err != nil {
		fmt.Fprintln(os.Stderr, "BootWhy: could not read blame data:", err)
		os.Exit(2)
	}
	journalCutoff := times.TargetAfter
	if journalCutoff > 0 {
		journalCutoff = times.Kernel + times.Initrd + times.TargetAfter + 2*time.Second
	} else if times.Userspace > 0 {
		journalCutoff = times.Kernel + times.Initrd + times.Userspace + 2*time.Second
	}
	journal, _, err := parse.ParseJournal(bytes.NewReader(raw.Journal), journalCutoff)
	if err != nil {
		fmt.Fprintln(os.Stderr, "BootWhy: could not read journal data:", err)
		os.Exit(2)
	}
	result := analyze.Analyze(times, chain, blame, raw.Ctx)
	findings := diagnose.Diagnose(result.Contributors, chain, journal, raw.Ctx, rules.Default())
	if command == "report" && len(flags.Args()) > 0 && flags.Args()[0] == "explain" {
		command = "explain"
	}
	if *jsonOutput {
		if err := report.JSON(os.Stdout, times, result, findings); err != nil {
			fmt.Fprintln(os.Stderr, "BootWhy: could not write JSON:", err)
			os.Exit(2)
		}
	} else if command == "explain" {
		explainArgs := flags.Args()
		if len(explainArgs) > 0 && explainArgs[0] == "explain" {
			explainArgs = explainArgs[1:]
		}
		if len(explainArgs) != 1 {
			fmt.Fprintln(os.Stderr, "BootWhy: usage: bootwhy explain <row>")
			os.Exit(2)
		}
		row, err := strconv.Atoi(explainArgs[0])
		if err != nil || row < 1 || row > len(findings) {
			fmt.Fprintln(os.Stderr, "BootWhy: no such finding")
			os.Exit(2)
		}
		report.Explain(os.Stdout, findings[row-1])
		return
	} else {
		report.Table(os.Stdout, times, result, findings, "systemd-analyze and journalctl")
		report.Timeline(os.Stdout, times, chain)
	}
	if len(findings) > 0 {
		os.Exit(1)
	}
}

func printReport(writer io.Writer, times parse.Times, chain []parse.ChainNode, result analyze.Result) {
	fmt.Fprintln(writer, "Boot phases:")
	table := tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	fmt.Fprintf(table, "Firmware:\t%s\n", times.Firmware)
	fmt.Fprintf(table, "Loader:\t%s\n", times.Loader)
	fmt.Fprintf(table, "Kernel:\t%s\n", times.Kernel)
	fmt.Fprintf(table, "Initrd:\t%s\n", times.Initrd)
	fmt.Fprintf(table, "Userspace:\t%s\n", times.Userspace)
	fmt.Fprintf(table, "Total settled:\t%s\n", times.Total)
	if times.Target != "" {
		fmt.Fprintf(table, "%s reached:\t%s in userspace\n", times.Target, times.TargetAfter)
	}
	table.Flush()

	fmt.Fprintln(writer, "\nCritical chain:")
	if len(chain) == 0 {
		fmt.Fprintln(writer, "  unavailable")
		return
	}
	for _, node := range chain {
		fmt.Fprintf(writer, "%s%s @%s", spaces(node.Depth), node.Unit, node.At)
		if node.Took > 0 {
			fmt.Fprintf(writer, " +%s", node.Took)
		}
		fmt.Fprintln(writer)
	}
	if len(result.Contributors) > 0 {
		fmt.Fprintln(writer, "\nLikely contributors:")
		for _, contributor := range result.Contributors {
			fmt.Fprintf(writer, "  %s: +%s (%.0f%%)\n", contributor.Unit, contributor.Added, contributor.Share*100)
		}
	}
}

func spaces(depth int) string {
	return strings.Repeat("  ", depth)
}
