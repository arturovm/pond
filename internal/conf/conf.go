package conf

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
)

// Basic settings.
var (
	Addr          string
	Port          int
	dataDir string
	Debug         bool
	Help          bool
)

func init() {
	flag.StringVarP(&Addr, "addr", "a", "127.0.0.1", "API server address")
	flag.IntVarP(&Port, "port", "p", 8080, "API server port")
	flag.StringVar(&dataDir, "datadir", "", "Pond data directory")
	flag.BoolVarP(&Debug, "debug", "d", false, "Enable debug mode")
	flag.BoolVarP(&Help, "help", "h", false, "Print this message")
	flag.Parse()
}

// PrintHelp prints flags and usage to standard output.
func PrintHelp() {
	fmt.Println("Usage: pond [options]")
	flag.CommandLine.SortFlags = false
	flag.PrintDefaults()
}

func binDir() string {
	binDir, err := os.Executable()
	if err != nil {
		slog.Warn("failed to obtain executable directory", "error", err)
	}

	binDir, err = filepath.EvalSymlinks(binDir)
	if err != nil {
		slog.Warn("failed to evaluate symlinks", "error", err)
	}

	return filepath.Dir(binDir)
}

// DataDir returns the application's data directory. If unset by the user it
// defaults to a location with the path "data" within the same directory as the
// executable.
func DataDir() string {
	if dataDir != "" {
		return dataDir
	}
	return filepath.Join(binDir(), "data")
}

