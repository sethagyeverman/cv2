// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"cv2/internal/pkg/response"
	"flag"
	"fmt"

	"cv2/internal/config"
	"cv2/internal/handler"
	"cv2/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/cv2.yaml", "the config file")

func main() {
	flag.Parse()

	logx.DisableStat()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	httpx.SetOkHandler(response.Success)
	httpx.SetErrorHandlerCtx(response.Error)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
