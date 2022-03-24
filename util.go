package create_app

import (
	"flag"
	"fmt"
	"github.com/moby/term"
	"github.com/spf13/pflag"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)


// TerminalSize returns the current width and height of the user's terminal. If it isn't a terminal,
// nil is returned. On error, zero values are returned for width and height.
// Usually w must be the stdout of the process. Stderr won't work.
func TerminalSize(w io.Writer) (int, int, error) {
	outFd, isTerminal := term.GetFdInfo(w)
	if !isTerminal {
		return 0, 0, fmt.Errorf("given writer is no terminal")
	}
	winsize, err := term.GetWinsize(outFd)
	if err != nil {
		return 0, 0, err
	}
	return int(winsize.Width), int(winsize.Height), nil
}

// HomeDir returns the home directory for the current user.
// On Windows:
// 1. the first of %HOME%, %HOMEDRIVE%%HOMEPATH%, %USERPROFILE% containing a `.apimachinery\config` file is returned.
// 2. if none of those locations contain a `.apimachinery\config` file, the first of
// %HOME%, %USERPROFILE%, %HOMEDRIVE%%HOMEPATH% that exists and is writeable is returned.
// 3. if none of those locations are writeable, the first of %HOME%, %USERPROFILE%,
// %HOMEDRIVE%%HOMEPATH% that exists is returned.
// 4. if none of those locations exists, the first of %HOME%, %USERPROFILE%,
// %HOMEDRIVE%%HOMEPATH% that is set is returned.
func HomeDir() string {
	if runtime.GOOS != "windows" {
		return os.Getenv("HOME")
	}
	home := os.Getenv("HOME")
	homeDriveHomePath := ""
	if homeDrive, homePath := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); len(homeDrive) > 0 && len(homePath) > 0 {
		homeDriveHomePath = homeDrive + homePath
	}
	userProfile := os.Getenv("USERPROFILE")

	// Return first of %HOME%, %HOMEDRIVE%/%HOMEPATH%, %USERPROFILE% that contains a `.apimachinery\config` file.
	// %HOMEDRIVE%/%HOMEPATH% is preferred over %USERPROFILE% for backwards-compatibility.
	for _, p := range []string{home, homeDriveHomePath, userProfile} {
		if len(p) == 0 {
			continue
		}
		if _, err := os.Stat(filepath.Join(p, ".apimachinery", "config")); err != nil {
			continue
		}
		return p
	}

	firstSetPath := ""
	firstExistingPath := ""

	// Prefer %USERPROFILE% over %HOMEDRIVE%/%HOMEPATH% for compatibility with other auth-writing tools
	for _, p := range []string{home, userProfile, homeDriveHomePath} {
		if len(p) == 0 {
			continue
		}
		if len(firstSetPath) == 0 {
			// remember the first path that is set
			firstSetPath = p
		}
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if len(firstExistingPath) == 0 {
			// remember the first path that exists
			firstExistingPath = p
		}
		if info.IsDir() && info.Mode().Perm()&(1<<(uint(7))) != 0 {
			// return first path that is writeable
			return p
		}
	}

	// If none are writeable, return first location that exists
	if len(firstExistingPath) > 0 {
		return firstExistingPath
	}

	// If none exist, return first location that is set
	if len(firstSetPath) > 0 {
		return firstSetPath
	}

	// We've got nothing
	return ""
}

// AddGlobalFlags explicitly registers flags that libraries (log, verflag, etc.) register
// against the global flagsets from "flag".
// We do this in order to prevent unwanted flags from leaking into the component's flagset.
func AddGlobalFlags(fs *pflag.FlagSet, name string) {
	fs.BoolP("help", "h", false, fmt.Sprintf("help for %s", name))
}

// normalize replaces underscores with hyphens
// we should always use hyphens instead of underscores when registering component flags.
func normalize(s string) string {
	return strings.ReplaceAll(s, "_", "-")
}

// Register adds a flag to local that targets the Value associated with the Flag named globalName in flag.CommandLine.
func Register(local *pflag.FlagSet, globalName string) {
	if f := flag.CommandLine.Lookup(globalName); f != nil {
		pflagFlag := pflag.PFlagFromGoFlag(f)
		pflagFlag.Name = normalize(pflagFlag.Name)
		local.AddFlag(pflagFlag)
	} else {
		panic(fmt.Sprintf("failed to find flag in global flagset (flag): %s", globalName))
	}
}
