package redirect

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/oschwald/geoip2-golang"
	"github.com/vulcand/oxy/utils"
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
	next       http.Handler
	from       string
	to         string
	status     int
	country    []string
	language   []string
	errHandler utils.ErrorHandler
	name       string
}

// New creates a geoRedirect middleware
func newGeoRedirect(next http.Handler, from string, to string, status int, country []string, language []string, name string) (http.Handler, error) {
	return &geoRedirect{
		next:       next,
		from:       from,
		to:         to,
		status:     status,
		country:    country,
		language:   language,
		errHandler: utils.DefaultHandler,
		name:       name,
	}, nil
}

func (r *geoRedirect) GetTracingInformation() (string, ext.SpanKindEnum) {
	return r.name, tracing.SpanKindNoneEnum
}

func (r *geoRedirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	return
}

func GetLanguage(req *http.Request) []string {
	language := req.Header.Get("Accept-Language")
	if language == "" {
		return []string{}
	}
	code := strings.Split(language, ";")[0]
	return strings.Split(code, ",")
}

func GetIp(str, dir string) string {
	db, err := geoip2.Open(dir)
	if err != nil {
		return "", err
	}
	defer db.Close()
	ip := net.ParseIP(str)
	record, err := db.City(ip)
	if err != nil {
		return "", err
	}
	return record.Country.IsoCode
}
