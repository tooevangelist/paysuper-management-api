package http

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/entrypoint"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	_ "github.com/micro/go-plugins/broker/rabbitmq"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	_ "github.com/micro/go-plugins/transport/grpc"
	"github.com/paysuper/paysuper-management-api/cmd"
	"github.com/paysuper/paysuper-management-api/internal/casbin"
	"github.com/paysuper/paysuper-management-api/internal/daemon"
	"github.com/paysuper/paysuper-management-api/pkg/http"
	"github.com/paysuper/paysuper-management-api/pkg/micro"
	"github.com/spf13/cobra"
	"sync"
)

var (
	casbinFlag bool
	Cmd    = &cobra.Command{
		Use:           "http",
		Short:         "HTTP API daemon",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			var (
				sHttp     *http.HTTP
				sMicro    *micro.Micro
				sCabin    *casbin.Casbin
				c         func()
				e         error
				ctxAll    context.Context
				ctxCancel context.CancelFunc
			)
			defer func() {
				if c != nil {
					c()
				}
			}()
			cmd.Slave.Executor(func(ctx context.Context) error {
				initial, _ := entrypoint.CtxExtractInitial(ctx)
				ctxAll, ctxCancel = context.WithCancel(ctx)
				if casbinFlag {
					sCabin, c, e = casbin.Build(ctxAll, initial, cmd.Observer)
					if e != nil {
						return e
					}
				}
				sHttp, c, e = daemon.BuildHTTP(ctxAll, initial, cmd.Observer)
				if e != nil {
					return e
				}
				sMicro, c, e = daemon.BuildMicro(ctxAll, initial, cmd.Observer)
				if e != nil {
					return e
				}
				return nil
			}, func(ctx context.Context) error {
				if casbinFlag {
					e := sCabin.ImportPolicy(cmd.Slave.WorkDir() + "/assets/policy.conf")
					if e != nil {
						sCabin.L().Error("import policy failed: %v", logger.Args(e.Error()))
						return e
					}
					sCabin.L().Info("casbin policy successfully applied")
				}
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					if err := sHttp.ListenAndServe(); err != nil {
						e = err
						ctxCancel()
					}
					wg.Done()
				}()
				go func() {
					if err := sMicro.ListenAndServe(); err != nil {
						e = err
						ctxCancel()
					}
					wg.Done()
				}()
				wg.Wait()
				return e
			})
		},
	}
)

func init() {
	// pflags
	Cmd.PersistentFlags().StringP(http.UnmarshalKeyBind, "b", ":0000", "bind address")
	Cmd.PersistentFlags().BoolVar(&casbinFlag, "casbin", false, "apply policy to casbin server")
}
