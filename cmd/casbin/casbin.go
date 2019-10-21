package casbin

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/entrypoint"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	_ "github.com/micro/go-plugins/broker/rabbitmq"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	_ "github.com/micro/go-plugins/transport/grpc"
	"github.com/paysuper/paysuper-management-api/cmd"
	"github.com/paysuper/paysuper-management-api/internal/casbin"
	"github.com/spf13/cobra"
	"os"
)

var (
	Cmd = &cobra.Command{
		Use:           "casbin",
		Short:         "Casbin policy migration",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(_ *cobra.Command, _ []string) {
			var (
				srv *casbin.Casbin
				c   func()
				e   error
			)
			defer func() {
				if c != nil {
					c()
				}
			}()
			cmd.Slave.Executor(func(ctx context.Context) error {
				initial, _ := entrypoint.CtxExtractInitial(ctx)
				srv, c, e = casbin.Build(ctx, initial, cmd.Observer)
				if e != nil {
					return e
				}
				return nil
			}, func(ctx context.Context) error {
				e := srv.ImportPolicy(cmd.Slave.WorkDir() + "/assets/policy.conf")
				if e != nil {
					srv.L().Error("import policy failed: %v", logger.Args(e.Error()))
					os.Exit(1)
				}
				return nil
			})
		},
	}
)
