package rpc

import (
	"fmt"
	"net/http"
	"text/template"
)

const debugText = `<html>
	<body>
	<title>GeeRPC Services</title>
	{{range .}}
	<hr>
	Service {{.Name}}
	<hr>
		<table>
		<th align=center>Method</th><th align=center>Calls</th>
		{{range $name, $mt := .Method}}
			<tr>
			<td align=left font=fixed>{{$name}}({{$mt.ArgType}}, {{$mt.ReplyType}}) error</td>
			<td align=center>{{$mt.NumCalls}}</td>
			</tr>
		{{end}}
		</table>
	{{end}}
	</body>
	</html>`

var debug = template.Must(template.New("RPC debug").Parse(debugText))

type debugHTTP struct {
	*Server
}

type debugService struct {
	Name   string
	Method map[string]*methodType
}

// Runs at /debug/rpc
func (server debugHTTP) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	// Build a sorted version of the data.
	var services []debugService
	server.serviceMap.Range(func(nameIf, svcIf interface{}) bool {
		svc := svcIf.(*service)
		services = append(services, debugService{
			Name:   nameIf.(string),
			Method: svc.method,
		})
		return true
	})
	err := debug.Execute(w, services)
	if err != nil {
		_, _ = fmt.Fprintln(w, "rpc: error executing template:", err.Error())
	}
}
