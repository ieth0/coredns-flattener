package flattener

import (
	"context"
	"fmt"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("flattener")

// CName is an example plugin to show how to write a plugin.
type CName struct {
	Next     plugin.Handler
	MaxDepth int
	Forward  HandlerWithCallbacks
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (c CName) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	nw := nonwriter.New(w)

	original := r.Copy()

	rCode, err := plugin.NextOrFailure(c.Name(), c.Next, ctx, nw, r)
	if err != nil {
		fmt.Println("Error Plugin Next Of Fail: ", err)
		return rCode, err
	}

	r = nw.Msg
	if r == nil {
		fmt.Println("Error no answer received")
		return 1, fmt.Errorf("no answer received")
	}

	if r.Answer != nil && len(r.Answer) > 0 && r.Answer[0].Header().Rrtype == dns.TypeCNAME {
		log.Debugf("flattening CNAME for request: %+v", r)
		// Follow the CNAME chain
		//fmt.Println(original.Question)
		//cNameAnswer := r.Answer[0].(*dns.CNAME).Target
		//newQuestion := new(dns.Question)
		//newQuestion.Name = cNameAnswer
		//newQuestion.Qtype = dns.TypeA // Change this to the desired record type
		//original.Answer = append(original.Answer, r.Answer[0])
		//original.Question = []dns.Question{*newQuestion}
		return c.Forward.ServeDNS(ctx, w, original)
	} else {
		log.Debug("Request didn't contain any answer or no CNAME")
	}

	err = w.WriteMsg(r)
	if err != nil {
		fmt.Println("Error write message", err)
		return 1, err
	}

	return 0, nil
}

// Name implements the Handler interface.
func (c CName) Name() string { return "flattener" }

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("flattener")
	return r.ResponseWriter.WriteMsg(res)
}

type HandlerWithCallbacks interface {
	plugin.Handler
	OnStartup() error
	OnShutdown() error
}
