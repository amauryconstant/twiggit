// Path: internal/cmdutil/factory.go
package cmdutil

import (
	"sync"

	"github.com/you/myapp/internal/config"
	"github.com/you/myapp/internal/git"
	"github.com/you/myapp/internal/iostreams"
)

// Factory hands commands their shared dependencies. Function fields are lazy:
// a dependency is built only when a command calls it inside RunE, so
// `myapp version` never reads config. Only cmd/ sees the Factory; commands
// pass plain values on to internal/core.
type Factory struct {
	IOStreams *iostreams.IOStreams
	Config    func() (*config.AppConfig, error)
	Client    func() (*git.Client, error)
}

func NewFactory() *Factory {
	f := &Factory{IOStreams: iostreams.System()}

	// sync.OnceValues caches the first result, so chained dependencies
	// (Client → Config) read the config file once per invocation.
	f.Config = sync.OnceValues(config.Load)
	f.Client = sync.OnceValues(func() (*git.Client, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		return git.NewClient(cfg.Timeout), nil
	})
	return f
}
