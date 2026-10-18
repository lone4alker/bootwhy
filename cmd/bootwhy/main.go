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
	"strings"
	"text/tabwriter"

	"github.com/lone4alker/bootwhy/internal/acquire"
	"github.com/lone4alker/bootwhy/internal/parse"
)

var version = "dev"

func main() {
	flags := flag.NewFlagSet("bootwhy", flag.ExitOnError)
	from := flags.String("from", "", "read a captured boot from DIR")
	showVersion := flags.Bool("version", false, "print version")
	flags.Parse(os.Args[1:])

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
	printReport(os.Stdout, times, chain)
}

func printReport(writer io.Writer, times parse.Times, chain []parse.ChainNode) {
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
}

func spaces(depth int) string {
	return strings.Repeat("  ", depth)
}
