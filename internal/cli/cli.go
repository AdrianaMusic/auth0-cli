package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/auth0/auth0-cli/internal/api"
	"github.com/auth0/auth0-cli/internal/display"
	"github.com/cyx/auth0/management"
	"github.com/spf13/cobra"
)

type data struct {
	DefaultTenant string            `json:"default_tenant"`
	Tenants       map[string]tenant `json:"tenants"`
}

type tenant struct {
	Domain string `json:"domain"`

	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`

	// TODO(cyx): This will be what we do with device flow.
	BearerToken string `json:"bearer_token,omitempty"`
}

type cli struct {
	api      *api.API
	renderer *display.Renderer

	verbose bool
	tenant  string

	initOnce sync.Once
	path     string
	data     data
}

func (c *cli) setup() error {
	t, err := c.getTenant()
	if err != nil {
		return err
	}

	var mgmt *management.Management
	if t.BearerToken != "" {
		mgmt, err = management.New(t.Domain,
			management.WithStaticToken(t.BearerToken),
			management.WithDebug(c.verbose))
	} else {
		mgmt, err = management.New(t.Domain,
			management.WithClientCredentials(t.ClientID, t.ClientSecret),
			management.WithDebug(c.verbose))
	}

	c.api = api.New(mgmt)
	return err
}

func (c *cli) getTenant() (tenant, error) {
	if err := c.init(); err != nil {
		return tenant{}, err
	}

	t, ok := c.data.Tenants[c.tenant]
	if !ok {
		return tenant{}, fmt.Errorf("Unable to find tenant: %s", c.tenant)
	}

	return t, nil
}

func (c *cli) init() error {
	var err error
	c.initOnce.Do(func() {
		if c.path == "" {
			c.path = path.Join(os.Getenv("HOME"), ".config", "auth0", "config.json")
		}

		var buf []byte
		if buf, err = ioutil.ReadFile(c.path); err != nil {
			return
		}

		if err = json.Unmarshal(buf, &c.data); err != nil {
			return
		}

		if c.tenant == "" && c.data.DefaultTenant == "" {
			err = fmt.Errorf("Not yet configured. Try `auth0 login`.")
			return
		}

		if c.tenant == "" {
			c.tenant = c.data.DefaultTenant
		}

		c.renderer = &display.Renderer{
			Tenant: c.tenant,
			Writer: os.Stdout,
		}
	})

	return err
}

func mustRequireFlags(cmd *cobra.Command, flags ...string) {
	for _, f := range flags {
		if err := cmd.MarkFlagRequired(f); err != nil {
			panic(err)
		}
	}
}
