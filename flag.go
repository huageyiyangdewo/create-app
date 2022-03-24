package create_app

import (
	goflag "flag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"strings"
)

//nolint:deadcode,unused,varcheck
func initFlag()  {
	pflag.CommandLine.SetNormalizeFunc(WordSepNormalizeFunc)
}

// WordSepNormalizeFunc changes all flags that contain "_" separators
func WordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
	}

	return pflag.NormalizedName(name)
}


// InitFlags normalizes, parses, then logs the command line flags.
func InitFlags(flags *pflag.FlagSet)  {
	flags.SetNormalizeFunc(WordSepNormalizeFunc)
	flags.AddGoFlagSet(goflag.CommandLine)
}

func PrintFlags(flags *pflag.FlagSet)  {
	flags.VisitAll(func(flag *pflag.Flag) {
		logrus.Debugf("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}