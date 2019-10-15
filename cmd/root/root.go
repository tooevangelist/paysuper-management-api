package root

import (
	"context"
	"fmt"
	"github.com/ProtocolONE/go-core/v2/pkg/config"
	"github.com/ProtocolONE/go-core/v2/pkg/entrypoint"
	"github.com/ProtocolONE/go-core/v2/pkg/invoker"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/alexeyco/simpletable"
	"github.com/fatih/color"
	"github.com/gurukami/typ/v2"
	"github.com/paysuper/paysuper-management-api/cmd"
	"github.com/paysuper/paysuper-management-api/cmd/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	configFile    string
	debug         bool
	gracefulDelay time.Duration
	initial       = config.Initial{}
	log           logger.Logger
	ep            entrypoint.Master
	e             error
	c             func()
)

const (
	prefix               = "cmd.root"
	envPrefix            = ""
	envWorkDir           = "WD"
	viperCfgType         = "yaml"
	defaultGracefulDelay = 50 * time.Millisecond
	allowEmptyEnv        = true
)

// http://www.patorjk.com/software/taag/#p=display&f=Calvin%20S&t=PaySuper%20M-API

var logo = `
╔═╗┌─┐┬ ┬╔═╗┬ ┬┌─┐┌─┐┬─┐  ╔╦╗   ╔═╗╔═╗╦
╠═╝├─┤└┬┘╚═╗│ │├─┘├┤ ├┬┘  ║║║───╠═╣╠═╝║
╩  ┴ ┴ ┴ ╚═╝└─┘┴  └─┘┴└─  ╩ ╩   ╩ ╩╩  ╩
	VERSION: %v`

// Root command
var rootCmd = &cobra.Command{
	Use:           "bin [command]",
	Long:          "",
	Short:         fmt.Sprintf(logo, version.Version()),
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(subCmd *cobra.Command, _ []string) error {

		// initializing
		initial.WorkDir = os.Getenv(envWorkDir)
		if len(initial.WorkDir) == 0 {
			dir, err := os.Getwd()
			if err == nil {
				initial.WorkDir = dir
			} else {
				initial.WorkDir, e = filepath.Abs(filepath.Dir(os.Args[0]))
				if e != nil {
					return e
				}
			}
		}
		initial.WorkDir, e = filepath.Abs(initial.WorkDir)
		if e != nil {
			return e
		}

		// bin pflags to viper
		e = initial.Viper.BindPFlags(subCmd.Parent().PersistentFlags())
		if e != nil {
			return e
		}
		e = initial.Viper.BindPFlags(subCmd.PersistentFlags())
		if e != nil {
			return e
		}

		initial.Viper.SetConfigFile(configFile)

		if configFile != "" {
			e := initial.Viper.ReadInConfig()
			if e != nil {
				return fmt.Errorf("can't read config, %v", errors.WithMessage(e, prefix))
			}
		}

		inv := invoker.NewInvoker()
		cmd.Observer = inv

		ep, c, e = entrypoint.Build(context.Background(), initial, inv)
		if e != nil {
			return e
		}
		defer c()

		cmd.Slave = ep
		log = ep.Logger().WithFields(logger.Fields{"service": prefix})

		go func() {
			reloadSignal := make(chan os.Signal, 1)
			signal.Notify(reloadSignal, syscall.SIGHUP)
			for {
				sig := <-reloadSignal
				inv.Reload(context.Background())
				ep.Logger().Info("OS signaled `%v`, reload", logger.Args(sig.String()))
			}
		}()

		go func() {
			shutdownSignal := make(chan os.Signal, 1)
			signal.Notify(shutdownSignal, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
			sig := <-shutdownSignal
			ep.Logger().Info("OS signaled `%v`, graceful shutdown in %s", logger.Args(sig.String(), gracefulDelay))
			ctx, _ := context.WithTimeout(context.Background(), gracefulDelay)
			ep.Shutdown(ctx, 0)
		}()

		return nil
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {

		preRun := func() error {
			if debug {
				fmt.Printf(logo, version.Version())
				fmt.Println(color.RedString("\n\n# DEBUG INFO\n"))
				fmt.Printf("\nWork directory: %v\n\n", ep.WorkDir())
				fmt.Println(color.GreenString("# CONFIG FILE SETTINGS\n\n"))

				var data [][]interface{}

				for _, item := range initial.Viper.AllEnrichedSettings() {
					data = append(data, []interface{}{
						item.Key + "\n-> " + strings.Join(item.ENV, "\n-> "),
						item.Value,
						item.Type,
					},[]interface{}{"","",""})
				}

				table := simpletable.New()

				table.Header = &simpletable.Header{
					Cells: []*simpletable.Cell{
						{Align: simpletable.AlignCenter, Text: "Key Path -> ENV"},
						{Align: simpletable.AlignCenter, Text: "Value"},
						{Align: simpletable.AlignCenter, Text: "Type"},
					},
				}

				for _, row := range data {
					r := []*simpletable.Cell{
						{Align: simpletable.AlignLeft, Text: typ.Of(row[0]).String().V()},
						{Align: simpletable.AlignLeft, Text: typ.Of(row[1]).String().V()},
						{Align: simpletable.AlignLeft, Text: typ.Of(row[2]).String().V()},
					}
					table.Body.Cells = append(table.Body.Cells, r)
				}

				table.SetStyle(simpletable.StyleMarkdown)
				fmt.Println(table.String())
				fmt.Println()

				fmt.Println(color.CyanString("\n# LOGS\n\n"))
			}

			_, err := maxprocs.Set(maxprocs.Logger(log.Printf))
			return err
		}

		if e = ep.Serve(preRun); e != nil {
			_ = preRun()
			log.Error(e.Error())
			ep.Shutdown(context.Background(), 0)
		}
	},
}

func init() {
	initial.Viper = config.NewViper()
	initial.Viper.SetConfigType(viperCfgType)
	initial.Viper.SetEnvPrefix(envPrefix)
	initial.Viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	initial.Viper.AutomaticEnv()
	initial.Viper.AllowEmptyEnv(allowEmptyEnv)
	// pflags
	rootCmd.PersistentFlags().StringVarP(&configFile, config.UnmarshalKeyConfigFile, "c", "", "config file")
	rootCmd.PersistentFlags().BoolVarP(&debug, config.UnmarshalKeyDebug, "d", false, "debug mode")
	rootCmd.PersistentFlags().StringP(logger.UnmarshalKeyLevel, "l", "info", "logger level")
	rootCmd.PersistentFlags().StringSliceP(logger.UnmarshalKeyDebugTags, "t", []string{}, "logger tags for filter output, e.g.: -t tag -t tag2 -t key:value")
	rootCmd.PersistentFlags().DurationVar(&gracefulDelay, config.UnmarshalKeyGracefulDelay, defaultGracefulDelay, "graceful delay")
}

func Execute(cmds ...*cobra.Command) {
	rootCmd.AddCommand(cmds...)
	if e := rootCmd.Execute(); e != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", e.Error())
		os.Exit(1)
	}
}

func ExecuteDefault(defaultArgs []string, cmds ...*cobra.Command) {
	if len(defaultArgs) != 0 {
		var cmdFromArgs string
		if len(os.Args) > 1 {
			cmdFromArgs = os.Args[1]
		}
		if strings.HasPrefix(cmdFromArgs, "-") || cmdFromArgs == "" {
			os.Args = append(os.Args[:1], append(defaultArgs, os.Args[1:]...)...)
		}
	}
	rootCmd.AddCommand(cmds...)
	if e := rootCmd.Execute(); e != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", e.Error())
		os.Exit(1)
	}
}
