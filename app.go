package create_app

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/marmotedu/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	progressMessage = color.GreenString("==>")

	// shell 脚本
	usageTemplate = fmt.Sprintf(`%s{{if .Runnable}}
  %s{{end}}{{if .HasAvailableSubCommands}}
  %s{{end}}{{if gt (len .Aliases) 0}}
%s
  {{.NameAndAliases}}{{end}}{{if .HasExample}}
%s
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}
%s{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  %s {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}
%s{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
Use "%s --help" for more information about a command.{{end}}
`,
		color.CyanString("Usage:"),
		color.GreenString("{{.UseLine}}"),
		color.GreenString("{{.CommandPath}} [command]"),
		color.CyanString("Aliases:"),
		color.CyanString("Examples:"),
		color.CyanString("Available Commands:"),
		color.GreenString("{{rpad .Name .NamePadding }}"),
		color.CyanString("Flags:"),
		color.CyanString("Global Flags:"),
		color.CyanString("Additional help topics:"),
		color.GreenString("{{.CommandPath}} [command]"),
	)
)


// App is the main structure of a cli application.
// It is recommended that an app be created with the app.NewApp() function.
type App struct {
	basename string
	name string
	description string
	options CliOptions
	runFunc RunFunc
	silence bool
	noVersion bool
	noConfig bool
	commands []*Command
	args cobra.PositionalArgs
	cmd *cobra.Command
}



// RunFunc defines the application's startup callback function
type RunFunc func(basename string) error


// WithRunFunc is used to set the application startup callback function option.
func WithRunFunc(run RunFunc) Option {
	return func(a *App) {
		a.runFunc = run
	}
}


// Option defines optional parameters for initializing the application structure.
type Option func(app *App)

// WithOptions to open the application's function to read from the command line
// or read parameters from the configuration file.
func WithOptions(opt CliOptions) Option {
	return func(a *App) {
		a.options = opt
	}
}

// WithDescription is used to set the description of the application.
func WithDescription(description string) Option {
	return func(a *App) {
		a.description = description
	}
}

// WithSilence sets the application to silent mode, in which the program startup
// information, configuration information, and version information are not
// printed in the console.
func WithSilence() Option {
	return func(a *App) {
		a.silence = true
	}
}

// WithNoVersion set the application does not provide version flag.
func WithNoVersion() Option {
	return func(a *App) {
		a.noVersion = true
	}
}

// WithNoConfig set the application does not provide config flag.
func WithNoConfig() Option {
	return func(a *App) {
		a.noConfig = true
	}
}


// WithValidArgs set the validation function to valid non-flag arguments.
func WithValidArgs(args cobra.PositionalArgs) Option {
	return func(a *App) {
		a.args = args
	}
}

// WithDefaultValidArgs set default validation function to valid non-flag arguments.
func WithDefaultValidArgs() Option {
	return func(a *App) {
		a.args = func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), arg)
				}
			}

			return nil
		}
	}
}

func NewApp(name, basename string, opts ...Option) *App {
	a := &App{
		name: name,
		basename: basename,
	}

	for _, opt := range opts {
		opt(a)
	}

	a.buildCommand()
	return a
}

func (a *App) buildCommand()  {
	cmd := cobra.Command{
		Use: FormatBasename(a.basename),
		Short: a.name,
		Long: a.description,
		// stop printing usage when the command errors
		SilenceUsage: true,
		SilenceErrors: true,
		Args: a.args,
	}

	//cmd.SetUsageTemplate(usageTemplate)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.Flags().SortFlags = true

	if len(a.commands) > 0 {
		for _, command := range a.commands {
			cmd.AddCommand(command.cobraCommand())
		}

		cmd.SetHelpCommand(helpCommand(a.name))
	}

	if a.runFunc != nil {
		cmd.RunE = a.runCommand
	}

	var namedFlagSets NamedFlagSets
	if a.options != nil {
		namedFlagSets = a.options.Flags()
		fs := cmd.Flags()
		for _, f := range namedFlagSets.FlagSets {
			fs.AddFlagSet(f)

		}

		usageFmt := "Usage:\n  %s\n"
		cols, _, _ := TerminalSize(cmd.OutOrStdout())
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
			PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
		})
		cmd.SetUsageFunc(func(cmd *cobra.Command) error {
			fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
			PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)

			return nil
		})

	}

	if a.noVersion {
		AddFlags(namedFlagSets.FlagSet("global"))
	}

	if !a.noConfig {
		addConfigFlag(a.basename, namedFlagSets.FlagSet("global"))
	}

	AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	a.cmd = &cmd
}


// Run is used to launch the application
func (a *App) Run()  {
	if err := a.cmd.Execute(); err != nil {
		fmt.Printf("%v %v\n", color.RedString("Error:", err))
		os.Exit(1)
	}
}

// Command return cobra command instance inside the application
func (a *App) Command() *cobra.Command {
	return a.cmd
}

func (a *App) runCommand(cmd *cobra.Command, args []string) error {
	printWorkDir()
	PrintFlags(cmd.Flags())

	if !a.noVersion {
		// display application version information
		PrintAndExitIfRequested()
	}

	if !a.noConfig {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		if err := viper.Unmarshal(a.options); err != nil {
			return err
		}
	}

	if !a.silence {
		logrus.Infof("%v Starting %s ...", progressMessage, a.name)
		if !a.noVersion {
			logrus.Infof("%v Version: `%s`", progressMessage, Get())
		}
		if !a.noConfig {
			logrus.Infof("%v Config file used: `%s`", progressMessage, viper.ConfigFileUsed())
		}
	}

	if a.options != nil {
		if err := a.applyOptionRules(); err != nil {
			return err
		}
	}

	if a.runFunc != nil {
		return a.runFunc(a.basename)
	}

	return nil
}

func (a *App) applyOptionRules() error {
	if completableOptions, ok := a.options.(CompletableOptions); ok {
		if err := completableOptions.Complete(); err != nil {
			return err
		}
	}

	if errs := a.options.Validate(); len(errs) != 0 {
		return errors.NewAggregate(errs)
	}

	if printableOptions, ok := a.options.(PrintableOptions); ok {
		logrus.Infof("%v Config: `%s`", progressMessage, printableOptions.String())
	}

	return nil
}

func printWorkDir()  {
	wd, _ := os.Getwd()
	logrus.Infof("%v WorkingDir: %s", progressMessage, wd)
}