package redirect

import (
	"context"
	"errors"
	"net/http"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
)

const (
	typeRedirectName = "RedirectGeoip"
)

// NewGeoRedirect creates a redirect middleware.
func NewGeoRedirect(ctx context.Context, next http.Handler, conf dynamic.RedirectGeoip, name string) (http.Handler, error) {
	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeRedirectName))
	logger.Debug("Creating middleware")
	logger.Debugf("Setting up redirection from %s to %s", conf.From, conf.To)
	if len(conf.From) == 0 || len(conf.To) {
		return nil, errors.New("you must provide 'from' and 'to'")
	}

	return newGeoRedirect(next, conf.From, conf.To, conf.Status, conf.Country, conf.Language)
}


type geoRedirect struct {
	next http.Handler
	from string
	to string
	status int
	country []string
	language []string
}

// New creates a geoRedirect middleware
func newGeoRedirect(next http.Handler, from string, to string, status int, country []string, language []string) (http.Handler, error) {
	return	
}

// type RedirectGeoip struct {
// 	From     string   `json:"from,omitempty" toml:"from,omitempty" yaml:"from,omitempty"`
// 	To       string   `json:"to,omitempty" toml:"to,omitempty" yaml:"to,omitempty"`
// 	Status   int      `json:"status,omitempty" toml:"status,omitempty" yaml:"status,omitempty"`
// 	Country  []string `json:"country,omitempty" toml:"country,omitempty" yaml:"country,omitempty"`
// 	Language []string `json:"language,omitempty" toml:"language,omitempty" yaml:"language,omitempty"`
// }