module github.com/memocash/server

replace github.com/gcash/bchd => ../../../pkg/github.com/gcash/bchd

go 1.16

require (
	github.com/99designs/gqlgen v0.14.0
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/gcash/bchd v0.17.1
	github.com/jchavannes/bchutil v0.0.0-20190601153950-94d7f52a5867
	github.com/jchavannes/btcd v0.0.0-20210319161304-69acdff53c2e
	github.com/jchavannes/go-mnemonic v0.0.0-20191017214729-76f026914b65
	github.com/jchavannes/jgo v0.0.0-20210920225626-5a88f5951c3c
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/tyler-smith/go-bip32 v1.0.0
	github.com/vektah/gqlparser/v2 v2.2.0
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)
