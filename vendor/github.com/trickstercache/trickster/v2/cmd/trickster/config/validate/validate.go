package validate

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/trickstercache/trickster/v2/cmd/trickster/config"
	"github.com/trickstercache/trickster/v2/pkg/cache"
	tl "github.com/trickstercache/trickster/v2/pkg/observability/logging"
	tr "github.com/trickstercache/trickster/v2/pkg/observability/tracing/registration"
	"github.com/trickstercache/trickster/v2/pkg/routing"

	"github.com/gorilla/mux"
)

func ValidateConfig(conf *config.Config) error {
	for _, w := range conf.LoaderWarnings {
		fmt.Println(w)
	}

	caches := make(map[string]cache.Cache)
	for k := range conf.Caches {
		caches[k] = nil
	}

	router := mux.NewRouter()
	mr := http.NewServeMux()
	logger := tl.ConsoleLogger(conf.Logging.LogLevel)

	tracers, err := tr.RegisterAll(conf, logger, true)
	if err != nil {
		return err
	}

	_, err = routing.RegisterProxyRoutes(conf, router, mr, caches, tracers, logger, true)
	if err != nil {
		return err
	}

	if conf.Frontend.TLSListenPort < 1 && conf.Frontend.ListenPort < 1 {
		return errors.New("no http or https listeners configured")
	}

	if conf.Frontend.ServeTLS && conf.Frontend.TLSListenPort > 0 {
		_, err = conf.TLSCertConfig()
		if err != nil {
			return err
		}
	}

	return nil
}
