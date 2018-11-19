package payment_system

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type PaymentSystem interface {
	CreatePayment()
	ProcessPayment()
}

type Settings struct {
	Url        string
	TerminalId string
	SecretKey  string
}

type Path struct {
	path   string
	method string
}

func LogRequest(req *http.Request, resp *http.Response)  {
	dump, err := httputil.DumpRequestOut(req, true)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%s\n\n", dump)

	dump, _ = httputil.DumpResponse(resp, true)
	fmt.Printf("%s\n\n", dump)
}