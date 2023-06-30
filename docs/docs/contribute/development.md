# Development Guide

Following are the steps to setup guardian development environment.

## Running locally

<details>
  <summary>Dependencies:</summary>

    - Git
    - Go 1.18 or above
    - PostgreSQL 13.2 or above

</details>

Clone the repo

```
git clone git@github.com:raystack/guardian.git
```

Install all the golang dependencies

```
make setup
```

Build guardian binary file

```
make build
```

Init server config. Customise with your local configurations.

```
make config
```

Run database migrations

```
./guardian server migrate -c config.yaml
```

Start guardian server

```
./guardian server start -c config.yaml
```

Initialise client configurations

```
./guardian config init
```

## Running tests

Running all unit tests

```sh
make test
```

Print code coverage

```
make coverage
```
