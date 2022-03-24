package create_app

import (
	"bytes"
	"fmt"
	"github.com/spf13/pflag"
	"io"
	"strings"
)

type NamedFlagSets struct {
	// Order is an ordered list of flag set names
	Order  []string

	// FlagSets stores the flag by name
	FlagSets map[string]*pflag.FlagSet
}


// FlagSet returns the flag set with the given name and adds it to the
// ordered name list if it is not in there yet.
func (nfs *NamedFlagSets) FlagSet(name string) *pflag.FlagSet {
	if nfs.FlagSets == nil {
		nfs.FlagSets = map[string]*pflag.FlagSet{}
	}

	if _, ok := nfs.FlagSets[name]; !ok {
		nfs.FlagSets[name] = pflag.NewFlagSet(name, pflag.ExitOnError)
		nfs.Order = append(nfs.Order, name)
	}

	return nfs.FlagSets[name]
}


// PrintSections prints the given names flag sets in sections, with the maximal given column number.
// if cols is zero, lines are not wrapped.
func PrintSections(w io.Writer, fss NamedFlagSets, cols int)  {
	for _, name := range fss.Order {
		fs := fss.FlagSets[name]
		if !fs.HasFlags() {
			continue
		}

		wideFS := pflag.NewFlagSet("", pflag.ExitOnError)
		wideFS.AddFlagSet(fs)

		var zzz string
		if cols > 24 {
			zzz = strings.Repeat("z", cols-24)
			wideFS.Int(zzz, 0, strings.Repeat("z", cols-24))
		}

		var buf bytes.Buffer
		fmt.Fprintf(&buf, "\n%s flags:\n%s", strings.ToUpper(name[:1]+name[:1]), wideFS.FlagUsagesWrapped(cols))

		if cols > 24 {
			i := strings.Index(buf.String(), zzz)
			lines := strings.Split(buf.String()[:i], "\n")
			fmt.Fprint(w, strings.Join(lines[:len(lines)-1], "\n"))
			fmt.Fprintln(w)
		} else {
			fmt.Fprint(w, buf.String())
		}
	}
}


// CliOptions abstracts configuration options for reading parameters from the
// command line.
type CliOptions interface {
	Flags() (fss NamedFlagSets)

	Validate() []error
}


// ConfigurationOptions abstracts configuration options for reading parameters
// from a configuration file.
type ConfigurationOptions interface {
	// ApplyFlags parsing parameters from the command line or configuration file to
	// the options instance
	ApplyFlags() []error
}

// CompletableOptions abstracts options which can be completed.
type CompletableOptions interface {
	Complete()  error
}


// PrintableOptions abstracts options which can be printed.
type PrintableOptions interface {
	String() string
}