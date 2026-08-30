package homeassistant

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/evcc-io/evcc/plugin/auth"
	"github.com/evcc-io/evcc/server/network"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/transport"
	"golang.org/x/oauth2"
)

// https://developers.home-assistant.io/docs/auth_api

func init() {
	auth.Register("homeassistant", NewHomeAssistantFromConfig)
}

func NewHomeAssistantFromConfig(other map[string]any) (oauth2.TokenSource, error) {
	var cc struct {
		URI      string
		Home     string // TODO remove deprecated
		Insecure bool
	}

	if err := util.DecodeOther(other, &cc); err != nil {
		return nil, err
	}

	uri := cc.URI

	if uri == "" && cc.Home != "" {
		uri = instanceUriByName(cc.Home)
		if uri == "" {
			return nil, fmt.Errorf("unknown instance: %s", cc.Home)
		}
	}

	return NewHomeAssistant(uri, cc.Insecure)
}

func NewHomeAssistant(uri string, insecure bool) (oauth2.TokenSource, error) {
	uri = strings.TrimRight(uri, "/") // normalize

	if source, ok := supervisorTokenSource(uri); ok {
		return source, nil
	}

	extUrl := network.Config().ExternalURL()
	redirectUri := extUrl + network.CallbackPath

	oc := oauth2.Config{
		ClientID:    extUrl,
		RedirectURL: redirectUri,
		Endpoint: oauth2.Endpoint{
			AuthURL:   uri + "/auth/authorize",
			TokenURL:  uri + "/auth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	// validate url
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	host := u.Host
	if h, _, err := net.SplitHostPort(u.Host); err == nil {
		host = h
	}

	// use instance name instead of host if discovered on mDNS
	if name := instanceNameByUri(uri); name != "" {
		host = name
	}

	log := util.NewLogger("homeassistant")
	ctx := util.WithLogger(context.Background(), log)

	if insecure {
		log.WARN.Println("insecure mode enabled - TLS certificate verification is disabled, use only for trusted local/self-signed instances")
		httpClient := &http.Client{Transport: transport.Insecure()}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	}

	return auth.NewOAuth(ctx, "HomeAssistant", host, &oc)
}

// supervisorTokenSource uses the token injected into a Home Assistant app.
// Keep this restricted to the Supervisor's internal Core proxy so the token
// can never be forwarded to a user-supplied or internet host.
func supervisorTokenSource(uri string) (oauth2.TokenSource, bool) {
	token := strings.TrimSpace(os.Getenv("SUPERVISOR_TOKEN"))
	if token == "" {
		return nil, false
	}

	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "http" || u.Hostname() != "supervisor" {
		return nil, false
	}
	if u.Path != "/core" && !strings.HasPrefix(u.Path, "/core/") {
		return nil, false
	}

	return oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
		TokenType:   "Bearer",
	}), true
}
