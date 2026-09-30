package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/visualize"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// serveCmd starts a local HTTP dashboard for the current workspace.
func serveCmdFactory() *cobra.Command {
	var (
		bind string
		port int
		open bool
	)
	c := &cobra.Command{
		Use:   "serve",
		Short: "启动本地 web dashboard（spec 进度 + drift）",
		Long: `启动一个本地 HTTP 服务，提供 free-kiro 项目状态的可视化 dashboard：

  $ free-kiro serve                       # 127.0.0.1:7373（默认 loopback）
  $ free-kiro serve --bind 0.0.0.0        # 暴露到局域网
  $ free-kiro serve --port 9000           # 自定义端口
  $ free-kiro serve --port 0              # 随机端口（启动时打印）
  $ free-kiro serve --open                # 启动后尝试打开浏览器（macOS / Linux desktop）

Endpoints:
  GET /              → 单页 SPA（vanilla JS，无构建）
  GET /api/summary   → JSON：所有 spec 状态 + drift
  GET /api/specs      → JSON：spec 列表
  GET /api/spec/<n>   → JSON：单个 spec 状态

Dashboard 每 5 秒自动 refresh（轮询 /api/summary）。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			holder, err := loadEngine()
			if err != nil {
				return err
			}
			ws := workspace.New(holder.ws.Root())
			addr := net.JoinHostPort(bind, fmt.Sprintf("%d", port))

			srv := visualize.NewServer(addr, ws, holder.engine())
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return exitWithError(ferrors.Wrap("serve.listen", err, fmt.Sprintf("listen %s", addr)))
			}

			fmt.Fprintf(cmd.OutOrStdout(), "free-kiro dashboard listening on %s\n", srv.URL())
			if open {
				openBrowser(srv.URL())
			}

			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			go func() {
				<-ctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				srv.Shutdown()
				_ = shutdownCtx
			}()

			if err := srv.ServeWith(ln); err != nil && ctx.Err() == nil {
				return exitWithError(err)
			}
			return nil
		},
	}
	c.Flags().StringVar(&bind, "bind", "127.0.0.1", "bind address (use 0.0.0.0 for LAN)")
	c.Flags().IntVar(&port, "port", 7373, "port (use 0 to pick a random free port)")
	c.Flags().BoolVar(&open, "open", false, "try to open the dashboard URL in a browser after start")
	return c
}

// openBrowser tries to launch the default browser. Best-effort.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	_ = cmd.Start()
}