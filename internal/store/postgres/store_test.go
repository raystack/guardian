package postgres_test

import (
	"fmt"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/raystack/guardian/internal/store"
	"github.com/raystack/guardian/internal/store/postgres"
	"github.com/raystack/salt/log"
)

var (
	storeConfig = store.Config{
		Host:     "localhost",
		User:     "test_user",
		Password: "test_pass",
		Name:     "test_db",
		SslMode:  "disable",
	}
)

func newTestStore(logger log.Logger) (*postgres.Store, *dockertest.Pool, *dockertest.Resource, error) {
	opts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		Env: []string{
			"POSTGRES_PASSWORD=" + storeConfig.Password,
			"POSTGRES_USER=" + storeConfig.User,
			"POSTGRES_DB=" + storeConfig.Name,
		},
	}

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not create dockertest pool: %w", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	storeConfig.Port = resource.GetPort("5432/tcp")

	// attach terminal logger to container if exists
	// for debugging purpose
	if logger.Level() == "debug" {
		logWaiter, err := pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
			Container:    resource.Container.ID,
			OutputStream: logger.Writer(),
			ErrorStream:  logger.Writer(),
			Stderr:       true,
			Stdout:       true,
			Stream:       true,
		})
		if err != nil {
			logger.Fatal("could not connect to postgres container log output", "error", err)
		}
		defer func() {
			err = logWaiter.Close()
			if err != nil {
				logger.Fatal("could not close container log", "error", err)
			}

			err = logWaiter.Wait()
			if err != nil {
				logger.Fatal("could not wait for container log to close", "error", err)
			}
		}()
	}

	// Tell docker to hard kill the container in 120 seconds
	if err := resource.Expire(120); err != nil {
		return nil, nil, nil, err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	var st *postgres.Store
	time.Sleep(5 * time.Second)
	if err = pool.Retry(func() error {
		st, err = postgres.NewStore(&storeConfig)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	err = setup(st)
	if err != nil {
		logger.Fatal("failed to setup and migrate DB", "error", err)
	}
	return st, pool, resource, nil
}

func purgeTestDocker(pool *dockertest.Pool, resource *dockertest.Resource) error {
	if err := pool.Purge(resource); err != nil {
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}

func setup(store *postgres.Store) error {
	var queries = []string{
		"DROP SCHEMA public CASCADE",
		"CREATE SCHEMA public",
	}
	for _, query := range queries {
		store.DB().Exec(query)
	}

	if err := store.Migrate(); err != nil {
		return err
	}

	return nil
}

func truncateTable(store *postgres.Store, tableName string) error {
	query := fmt.Sprintf(`TRUNCATE "%s" CASCADE;`, tableName)
	return store.DB().Exec(query).Error
}
