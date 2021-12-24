module github.com/odpf/guardian

go 1.15

require (
	cloud.google.com/go/bigquery v1.8.0
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/antonmedv/expr v1.9.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.6.0
	github.com/imdario/mergo v0.3.11
	github.com/mcuadros/go-defaults v1.2.0
	github.com/mcuadros/go-lookup v0.0.0-20200831155250-80f87a4fa5ee
	github.com/mitchellh/mapstructure v1.4.1
	github.com/odpf/salt v0.0.0-20211223232616-de875a7b94cb
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	google.golang.org/api v0.44.0
	google.golang.org/genproto v0.0.0-20210903162649-d08c68adba83
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/datatypes v1.0.0
	gorm.io/driver/postgres v1.0.8
	gorm.io/gorm v1.20.12
)
