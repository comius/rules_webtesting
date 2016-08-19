// Package environment provides an interface for defining how browser environments
// are managed.
package environment

import (
	"context"
	"sync"

	"github.com/web_test_launcher/launcher/errors"
	"github.com/web_test_launcher/launcher/healthreporter"
	"github.com/web_test_launcher/metadata/capabilities"
	"github.com/web_test_launcher/metadata/metadata"
)

// Env allows web_test environments to be started for controlling a browser
// using Selenium WebDriver.
type Env interface {
	healthreporter.HealthReporter
	// SetUp is called once at the beginning of the test run, and should start a
	// WebDriver server. It is not necessary that the environment be healthy when
	// this returns. capsFile is the location of capabilities that should be merged
	// client-specified capabilities when provisioning a browser.
	SetUp(ctx context.Context) error
	// StartSession is called for each new WebDriver session, before
	// the new session command is sent to the WebDriver server.
	// caps is the capabilities sent to the proxy from the client, and
	// the return value is the capabilities that should be actually
	// sent to the WebDriver server new session command.
	StartSession(ctx context.Context, id int, caps map[string]interface{}) (map[string]interface{}, error)
	// StartSession is called for each new WebDriver session, before
	// the delete session command is sent to the WebDriver server.
	StopSession(ctx context.Context, id int) error
	// TearDown is called at the end of the test run.
	TearDown(ctx context.Context) error
	// WDAddress returns the address of the WebDriver server.
	WDAddress(context.Context) string
}

// Base is a partial implementation of Env useful as the base struct for
// implementations of Env.
type Base struct {
	name     string
	metadata metadata.Metadata

	mu      sync.RWMutex
	started bool
	stopped bool
}

// NewBase creates a new Base environment with the given component name.
func NewBase(name string, m metadata.Metadata) (*Base, error) {
	return &Base{
		name:     name,
		metadata: m,
	}, nil
}

// SetUp starts the URLService.
func (b *Base) SetUp(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.started {
		return errors.NewPermanent(b.Name(), "already started")
	}
	b.started = true
	return nil
}

// StartSession merges the passed in caps with b.metadata.caps and returns the merged
// capabilities that should be used when calling new session on the WebDriver
// server.
func (b *Base) StartSession(ctx context.Context, id int, caps map[string]interface{}) (map[string]interface{}, error) {
	if err := b.Healthy(ctx); err != nil {
		return nil, err
	}
	updated := capabilities.Merge(b.metadata.Capabilities, caps)
	return updated, nil
}

// StopSession is a no-op implementation of Env.StopSession.
func (b *Base) StopSession(ctx context.Context, _ int) error {
	return b.Healthy(ctx)
}

// TearDown is a no-op implementation of Env.TearDown.
func (b *Base) TearDown(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.started {
		return errors.NewPermanent(b.Name(), "not been started")
	}
	if b.stopped {
		return errors.NewPermanent(b.Name(), "already stopped")
	}
	b.stopped = true
	return nil
}

// Healthy always returns nil.
func (b *Base) Healthy(context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.started {
		return errors.NewPermanent(b.Name(), "not been started")
	}
	if b.stopped {
		return errors.NewPermanent(b.Name(), "already stopped")
	}
	return nil
}

// WDAddress returns the empty string.
func (*Base) WDAddress(context.Context) string {
	return ""
}

// Name returns the name of this environment to be used in error messages.
func (b *Base) Name() string {
	return b.name
}
