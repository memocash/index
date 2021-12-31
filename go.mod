module github.com/memocash/index

replace (
	github.com/gcash/bchd => ../../../pkg/github.com/gcash/bchd
	github.com/jchavannes/bchutil => ../../jchavannes/bchutil
	github.com/jchavannes/btcd => ../../jchavannes/btcd
	github.com/jchavannes/btclog => ../../jchavannes/btclog
	github.com/jchavannes/btcutil => ../../jchavannes/btcutil
)

go 1.16

require (
	github.com/99designs/gqlgen v0.14.0
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/gcash/bchd v0.17.1
	github.com/jchavannes/bchutil v0.0.0-20190601153950-94d7f52a5867
	github.com/jchavannes/btcd v0.0.0-20211231102419-9f97c2166438
	github.com/jchavannes/btclog v0.0.0-20211231060513-6ff05f5c3d70
	github.com/jchavannes/btcutil v1.0.3-0.20211231102310-a1560bb282e9
	github.com/jchavannes/go-mnemonic v0.0.0-20191017214729-76f026914b65
	github.com/jchavannes/jgo v0.0.0-20211112043704-31caacec985a
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/tyler-smith/go-bip32 v1.0.0
	github.com/vektah/gqlparser/v2 v2.2.0
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)
