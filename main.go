package main

import (
	"github.com/paysuper/paysuper-management-api/cmd/casbin"
	"github.com/paysuper/paysuper-management-api/cmd/http"
	"github.com/paysuper/paysuper-management-api/cmd/root"
)

func main() {
	args := []string{
		"http", "-c", "configs/local.yaml", "-d",
	}
	root.ExecuteDefault(args, http.Cmd, casbin.Cmd)
}
