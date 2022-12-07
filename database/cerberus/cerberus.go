package cerberus

import (
	"context"
	"fmt"
	cerberus "github.com/a11n-io/go-cerberus"
	"github.com/golang-migrate/migrate/v4/database"
	"go.uber.org/atomic"
	"io"
	nurl "net/url"
)

func init() {
	database.Register("cerberus", &Cerberus{})
}

var (
	ErrDatabaseDirty = fmt.Errorf("database is dirty")
	ErrNilConfig     = fmt.Errorf("no config")
)

type Config struct {
}

type Cerberus struct {
	client   cerberus.CerberusClient
	isLocked atomic.Bool
	context  context.Context
	config   *Config
}

func WithInstance(instance cerberus.CerberusClient, config *Config) (database.Driver, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	ctx := context.Background()

	if err := instance.Ping(ctx); err != nil {
		return nil, err
	}

	cerberusTokenPair, err := instance.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	c := context.WithValue(ctx, "cerberusTokenPair", cerberusTokenPair)

	mx := &Cerberus{
		client:  instance,
		config:  config,
		context: c,
	}
	return mx, nil
}

func (m *Cerberus) Open(url string) (database.Driver, error) {
	uri, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}
	apiKey := uri.User.Username()
	apiSecret, _ := uri.User.Password()

	instance := cerberus.NewClient(url, apiKey, apiSecret)

	mx, err := WithInstance(instance, &Config{})
	if err != nil {
		return nil, err
	}
	return mx, nil
}

func (m *Cerberus) Close() error {
	return nil
}

func (m *Cerberus) Drop() (err error) {
	return nil
}

func (m *Cerberus) Lock() error {
	fmt.Println("Cerberus Lock")
	if !m.isLocked.CAS(false, true) {
		return database.ErrLocked
	}
	return nil
}

func (m *Cerberus) Unlock() error {
	fmt.Println("Cerberus Unlock")
	if !m.isLocked.CAS(true, false) {
		return database.ErrNotLocked
	}
	return nil
}

func (m *Cerberus) Run(migration io.Reader) error {
	migr, err := io.ReadAll(migration)
	if err != nil {
		return err
	}
	query := string(migr[:])
	return m.executeQuery(query)
}

func (m *Cerberus) executeQuery(query string) error {

	return m.client.RunScript(m.context, query)
}

func (m *Cerberus) SetVersion(version int, dirty bool) error {

	err := m.client.SetMigrationVersion(m.context, cerberus.MigrationVersion{
		Version: version,
		Dirty:   dirty,
	})

	if err != nil {
		return &database.Error{OrigErr: err, Err: "set version failed"}
	}

	return nil
}

func (m *Cerberus) Version() (int, bool, error) {
	migrationVersion, err := m.client.GetMigrationVersion(m.context)
	if err != nil {
		return database.NilVersion, false, nil
	}

	return migrationVersion.Version, migrationVersion.Dirty, nil
}
