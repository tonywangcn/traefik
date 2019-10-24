package redirect

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/oschwald/geoip2-golang"
	"github.com/vulcand/oxy/utils"
	"golang.org/x/text/language"
)

const (
	typeRedirectName = "RedirectGeoip"
)

var dir string
var db *geoip2.Reader
var err error

// NewGeoRedirect creates a redirect middleware.
func NewGeoRedirect(ctx context.Context, next http.Handler, conf dynamic.RedirectGeoip, name string) (http.Handler, error) {
	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeRedirectName))
	logger.Debug("Creating middleware")
	logger.Debugf("Setting up redirection from %s to %s", conf.From, conf.To)
	if len(conf.From) == 0 || len(conf.To) == 0 {
		return nil, errors.New("you must provide 'from' and 'to'")
	}
	dir = conf.Dir
	db, err = geoip2.Open(dir)
	if err != nil {
		return nil, errors.New("geoip db initialize failed")
	}
	defer db.Close()
	return newGeoRedirect(next, conf.From, conf.To, conf.Status, conf.Country, conf.Language, conf.Dir, name)
}

type geoRedirect struct {
	next       http.Handler
	from       string
	to         string
	status     int
	country    []string
	language   []string
	dir        string
	errHandler utils.ErrorHandler
	name       string
}

// New creates a geoRedirect middleware
func newGeoRedirect(next http.Handler, from string, to string, status int, country []string, language []string, dir string, name string) (http.Handler, error) {
	return &geoRedirect{
		next:       next,
		from:       from,
		to:         to,
		status:     status,
		country:    country,
		language:   language,
		dir:        dir,
		errHandler: utils.DefaultHandler,
		name:       name,
	}, nil
}

func (r *geoRedirect) GetTracingInformation() (string, ext.SpanKindEnum) {
	return r.name, tracing.SpanKindNoneEnum
}

func (r *geoRedirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// if country code in country or language code in language redirect
	parsedToURL, err := url.Parse(r.to)
	if err != nil {
		r.errHandler.ServeHTTP(rw, req, err)
		return
	}

	if req.URL.Path == r.from {
		userIP := GetUserIP(req)
		// fmt.Println("userIP", userIP)
		// countryCode := GetCountryCode("118.140.196.130", r.dir)
		countryCode := GetCountryCode(userIP, r.dir)
		languageCode := GetLanguage(req)
		isLanguage := inSlice(languageCode, r.language)
		isCountry := inSlice(countryCode, r.country)
		if isLanguage || isCountry || (len(r.language) == 0 && len(r.country) == 0) {
			handler := &redirectHandler{location: parsedToURL, status: r.status}
			handler.ServeHTTP(rw, req)
			return
		}

	}
	r.next.ServeHTTP(rw, req)
	return
}

type redirectHandler struct {
	location *url.URL
	status   int
}

func (r *redirectHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Location", r.location.String())
	status := http.StatusFound
	if req.Method != http.MethodGet {
		status = http.StatusTemporaryRedirect
	}
	if r.status == 301 {
		status = http.StatusMovedPermanently
		if req.Method != http.MethodGet {
			status = http.StatusPermanentRedirect
		}
	}
	rw.WriteHeader(status)
	_, err := rw.Write([]byte(http.StatusText(status)))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func GetUserIP(req *http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		ip, _, err = net.SplitHostPort(req.Header.Get("X-Real-Ip"))
		if err != nil {
			return ""
		}
		// return req.RemoteAddr
	}
	userIP := net.ParseIP(ip)
	if userIP == nil {
		return ""
	}
	return userIP.String()
}

func inSlice(source, target []string) bool {
	for _, s := range source {
		for _, t := range target {
			if strings.ToUpper(s) == strings.ToUpper(t) {
				return true
			}
		}
	}
	return false
}

func GetLanguage(req *http.Request) []string {
	lang, _ := req.Cookie("lang")
	accept := req.Header.Get("Accept-Language")
	t, _, _ := language.ParseAcceptLanguage(accept)
	if lang != nil {
		fmt.Println(lang)
		return []string{lang.String()}
	}
	tmpSlice := []string{}
	for _, v := range t {
		tmpSlice = append(tmpSlice, v.String())
	}
	if len(tmpSlice) > 4 {
		return tmpSlice[:4]
	}
	return tmpSlice
}

func GetCountryCode(str, dir string) []string {
	// db, err := geoip2.Open(dir)
	// if err != nil {
	// 	return []string{""}
	// }
	// defer db.Close()
	ip := net.ParseIP(str)
	record, err := db.City(ip)
	if err != nil {
		return []string{""}
	}
	return []string{record.Country.IsoCode}
}
