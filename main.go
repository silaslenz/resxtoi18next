package resxtoi18next

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("visitor_ip", parseCaddyfile)
}

// Middleware implements an HTTP handler that writes the
// visitor's IP address to a file or stream.
type Middleware struct {
	// The file or stream to write to. Can be "stdout"
	// or "stderr".
	Output string `json:"output,omitempty"`
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.visitor_ip",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error {
	return nil
}

// Validate implements caddy.Validator.
func (m *Middleware) Validate() error {
	return nil
}

type Users struct {
	XMLName xml.Name `xml:"root"`
	Users   []Data   `xml:"data"`
}

type Data struct {
	XMLName xml.Name `xml:"data"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value"`
}

type I18Next struct {
	string      `json:"birdType"`
	Description string `json:"what it does"`
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	xmlFile, err := os.Open(m.Output)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Successfully Opened %s", m.Output)
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	// we initialize our Users array
	var users Users
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &users)
	var jsonmap = make(map[string]interface{})

	for i := 0; i < len(users.Users); i++ {
		key := users.Users[i].Name
		keyparts := strings.Split(key, ".")
		var layer = jsonmap
		for index, element := range keyparts {
			if index == len(keyparts)-1 {
				layer[element] = users.Users[i].Value
				break
			}
			if layer[element] == nil {
				layer[element] = make(map[string]interface{})
				layer = layer[element].(map[string]interface{})
			}

		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonmap)
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.Args(&m.Output) {
			return d.ArgErr()
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
