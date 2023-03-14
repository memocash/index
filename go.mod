module github.com/memocash/index

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220314234659-1baeb1ce4c0b
	golang.org/x/net => golang.org/x/net v0.1.1-0.20221104162952-702349b0e862
	golang.org/x/text => golang.org/x/text v0.3.8
)

go 1.16

require (
	github.com/99designs/gqlgen v0.17.20
	github.com/jchavannes/bchutil v1.1.5-0.20220519214029-6a6c086b1f21
	github.com/jchavannes/btcd v1.1.5-0.20230112162803-412def37b600
	github.com/jchavannes/btclog v1.1.0
	github.com/jchavannes/btcutil v1.1.4
	github.com/jchavannes/go-mnemonic v0.0.0-20191017214729-76f026914b65
	github.com/jchavannes/jgo v0.0.0-20230222214331-95b230651774
	github.com/mattn/go-sqlite3 v1.14.16
	github.com/pkg/profile v1.6.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/tyler-smith/go-bip32 v1.0.0
	github.com/vektah/gqlparser/v2 v2.5.1
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.28.0
)
