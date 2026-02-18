package gen_test

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"log"
	"testing"
)

type MultiTest struct {
	Request  gen.MultiRequest
	Error    error
	TxHashes []test_tx.TxHash
}

func TestMulti(t *testing.T) {
	for j, multiTest := range multiTests {
		txs, err := gen.Multi(multiTest.Request)
		if testing.Verbose() {
			log.Printf("MultiTest %d:\n", j)
		}
		test_tx.Checker{
			Name:     fmt.Sprintf("TestMulti %d", j),
			Txs:      txs,
			Error:    multiTest.Error,
			TxHashes: multiTest.TxHashes,
		}.Check(err, t)
	}
}

// No inputs, no signatures - so still works
var multiTest0empty = MultiTest{
	Request: gen.MultiRequest{},
	Error:   gen.NotEnoughValueError,
}

var multiTest1like = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{&test_tx.LikeEmptyPostOutput},
		Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
		Change:  wallet.GetChange(test_tx.Address1),
		KeyRing: wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key1String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "985f3841672cfc1a61a547f9f98e0d618f4b5d90f4b6ed3d68597075c09ba9f4",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100ec3512456cdf33d5fb57ab7c46896966235d38507b4c89b009a58882013f126902203a3c5b963f28b2166bfcea3ed77a0eda58ba80bb4e1cc13d889c268e0addb0a9412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2b2850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}},
}

var multiTest2faucetLike = MultiTest{
	Request: gen.MultiRequest{
		Outputs:      []*memo.Output{&test_tx.LikeEmptyPostOutput},
		FaucetGetter: gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
		FaucetSaver:  test_tx.GetFaucetSaverWithKey(test_tx.Key1String),
		Change:       wallet.GetChange(test_tx.Address2),
		KeyRing:      wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key2String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "0dca339f4d3934393bd68ffcf171d8ff8447ec37f96c757a42eca3c7c0dfc941",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100c4a4d638e9ef67b363d7251618366a8be9d49e268d2abe0e34b74f43fe17a03102201d48b65dfd51d2d6cadbe484a75b1dc6a6f2527ca71a44537f269bf0e58210a2412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff022e260000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac905f0100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}, {
		TxHash: "cc67cfb4a2d106ccd1b6c7849c8f966be42cc3f268464f51c93429864ab4f797",
		TxRaw:  "010000000141c9dfc0c7a3ec427a756cf937ec4784ffd871f1fc8fd63b3934394d9f33ca0d000000006b483045022100aae7cb24ad6d10760c1542bd1dcd4f10154cf03c3f180c70e8d5dd8fb8df8dde022007a4e365a91267853ae6015b67f7aa898d6ab9899a201e62ff6b7fb548d7d4d1412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d240250000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}},
}

var multiTest3faucetEmpty = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{&test_tx.LikeEmptyPostOutput},
		Change:  wallet.GetChange(test_tx.Address2),
		KeyRing: wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key2String)),
	},
	Error: gen.NotEnoughValueError,
}

var multiTest4noFaucetChange = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{&test_tx.LikeEmptyPostOutput},
		FaucetGetter: gen.GetWrapper(&test_tx.TestFaucetGetter{TestGetter: test_tx.TestGetter{
			UTXOs: []memo.UTXO{test_tx.Address1InputUtxo8k},
		}}, test_tx.Address1pkHash),
		FaucetSaver: test_tx.GetFaucetSaverWithKey(test_tx.Key1String),
		Change:      wallet.GetChange(test_tx.Address2),
		KeyRing:     wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key2String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "29ec151ecb297b379b6a081990eed59df58495b8acf4bf517f7c41fdfe0297af",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a47304402200695123ed096a8327bf0e42db4c35e6b42a45ccff9749a4cfeb521e53c37d25c0220557c6224ae038ee2405f3ef7bcb7a443028a87d6129bd8fb300c526a536444ee412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff01801e0000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
	}, {
		TxHash: "29b1ce7ace8fe6d18e0e2e0282acc49c1b0c421312667ae9e849493cb56c82e8",
		TxRaw:  "0100000001af9702fefd417c7f51bff4acb89584f59dd5ee9019086a9b377b29cb1e15ec29000000006b483045022100c29c7da371e0dbfc3ad626f3b3f0eebc36c967212a208a477abd9a2d24a61da40220423a8ed7169d8ff9a9f46eaa2a42d1a7b92128c7c437789b0c9917cb2d664d43412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2921d0000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}},
}

var multiTest5faucetNotEnoughValue = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{&test_tx.LikeEmptyPostOutput},
		FaucetGetter: gen.GetWrapper(&test_tx.TestFaucetGetter{TestGetter: test_tx.TestGetter{
			UTXOs: []memo.UTXO{test_tx.Address1InputUtxo700},
		}}, test_tx.Address1pkHash),
		FaucetSaver: test_tx.GetFaucetSaverWithKey(test_tx.Key1String),
		Change:      wallet.GetChange(test_tx.Address2),
		KeyRing:     wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key2String)),
	},
	Error: gen.BelowDustLimitError,
}

var multiTest6maxSend = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{
			gen.GetAddressOutput(test_tx.Address2, memo.GetMaxSendForUTXOs(test_tx.UtxosAddress1twoRegular)),
		},
		Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: test_tx.UtxosAddress1twoRegular}, test_tx.Address1pkHash),
		Change:  wallet.GetChange(test_tx.Address1),
		KeyRing: wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key1String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "5979c0a890223f55c02800ce70f1eb92e1d4c7544257ada6a7e531937bb87bc6",
		TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a4730440220039bb07a587d1434fce162358e71d9e32202544c47fcc49dfaf58327c5e8f2e202203d3cc66d0ab1c694245d318b77d4ce8129b810b890876f4aaf82c5d5b5581314412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006a473044022055d333fb19d5a2f457a32ee3f5d3e20a404d2e1125c4d807219207cb68262219022015fba21f12507c9cf2bc4e3fe3a4b1b721984a241c9620f4db36e94fa1e8dfa5412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff019e080000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
	}},
}

var multiTest11maxSendWithToken = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{
			gen.GetAddressOutput(test_tx.Address2, memo.GetMaxSendForUTXOs(test_tx.UtxosAddress1twoRegularWithToken)),
		},
		Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: test_tx.UtxosAddress1twoRegularWithToken}, test_tx.Address1pkHash),
		Change:  wallet.GetChange(test_tx.Address1),
		KeyRing: wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key1String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "ec98a3cbe2750edae2e5cf21f94266bde30d3f94ca68d52880ded07da5c35f5c",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022055790d9d5cf5adc216eaac13fd075084a2a48b151029134643e3986c915cf270022078977e105a194863a5b545f30a468e84b0741c9749fda5565d7af26b1af3cb83412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0110070000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
	}},
}

var multiTest7sendTooMuchWithToken = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{
			gen.GetAddressOutput(test_tx.Address2, memo.GetMaxSendForUTXOs(test_tx.UtxosAddress1twoRegular)),
		},
		Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: test_tx.UtxosAddress1twoRegularWithToken}, test_tx.Address1pkHash),
		Change:  wallet.GetChange(test_tx.Address1),
		KeyRing: wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key1String)),
	},
	Error: gen.NotEnoughValueError,
}

var multiTest8faucetAndTokensLike = MultiTest{
	Request: gen.MultiRequest{
		Outputs:      []*memo.Output{&test_tx.LikeEmptyPostOutput},
		FaucetGetter: gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
		FaucetSaver:  test_tx.GetFaucetSaverWithKey(test_tx.Key1String),
		Getter:       gen.GetWrapper(&test_tx.TestGetter{UTXOs: test_tx.Address2InputsAll3Tokens}, test_tx.Address2pkHash),
		Change:       wallet.GetChange(test_tx.Address2),
		KeyRing:      wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key2String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "0dca339f4d3934393bd68ffcf171d8ff8447ec37f96c757a42eca3c7c0dfc941",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100c4a4d638e9ef67b363d7251618366a8be9d49e268d2abe0e34b74f43fe17a03102201d48b65dfd51d2d6cadbe484a75b1dc6a6f2527ca71a44537f269bf0e58210a2412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff022e260000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac905f0100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}, {
		TxHash: "cc67cfb4a2d106ccd1b6c7849c8f966be42cc3f268464f51c93429864ab4f797",
		TxRaw:  "010000000141c9dfc0c7a3ec427a756cf937ec4784ffd871f1fc8fd63b3934394d9f33ca0d000000006b483045022100aae7cb24ad6d10760c1542bd1dcd4f10154cf03c3f180c70e8d5dd8fb8df8dde022007a4e365a91267853ae6015b67f7aa898d6ab9899a201e62ff6b7fb548d7d4d1412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d240250000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}},
}

var multiTest9faucetNotEnoughValue = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{&test_tx.LikeEmptyPostOutput},
		FaucetGetter: gen.GetWrapper(&test_tx.TestFaucetGetter{TestGetter: test_tx.TestGetter{
			UTXOs: []memo.UTXO{test_tx.Address1InputUtxo1k},
		}}, test_tx.Address1pkHash),
		FaucetSaver: test_tx.GetFaucetSaverWithKey(test_tx.Key1String),
		Change:      wallet.GetChange(test_tx.Address2),
		KeyRing:     wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key2String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "78a9eee982e4f186af52e96d46faa5f18137427e1f6fe332cb3207074b25f598",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100d72c397b019195ab0d27a17e7a799e495dee48b417d98f5103fea48d247bc96f022020ca05deeea3a2c8e61fcc9b91465c854e2ae6b5377e1ee39355a8ceb74bd226412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0128030000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
	}, {
		TxHash: "1e807603d481ccbf02b7c49f3eaf5407d9956044d19bf9f3d345507f4e6acacf",
		TxRaw:  "010000000198f5254b070732cb32e36f1f7e423781f1a5fa466de952af86f1e482e9eea978000000006b483045022100ca8d0b6ad907f59c1f991fe863ab09f5aaafbbfae76330dd4db7ae67cfde158302201a78c901b4bb5075c485b474008453ed9239d27f9fb9f2f8d31a6ff2dbd83e0c412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d23a020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}},
}

var multiTest10faucetNewPost = MultiTest{
	Request: gen.MultiRequest{
		Outputs:      []*memo.Output{&test_tx.NewPostOutput},
		FaucetGetter: gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo1255}}, test_tx.Address1pkHash),
		FaucetSaver:  test_tx.GetFaucetSaverWithKey(test_tx.Key1String),
		Change:       wallet.GetChange(test_tx.Address2),
		KeyRing:      wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key2String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "0e16168d3ceeb8206e2cfb99efa4834a3a6c78a855331f3f5614c2738944e7eb",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100f549dd886c62587dd91b2107c21e48695e0bca619493a605f959463fefc8633b02207a8c9a66c37e9d40c1ece6245885e36226d5f9329006d04e433e71e35deeae59412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0127040000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
	}, {
		TxHash: "c0db892e9ee6a13934b3061a675c1af6791a79d8a5468f1b5a50c0d5bb3ecc89",
		TxRaw:  "0100000001ebe7448973c214563f1f3355a8786c3a4a83a4ef99fb2c6e20b8ee3c8d16160e000000006b483045022100a263b51649e01c8b629d6f9b7775d31aa433aef810977eb9f00f7230d7e7ecfc02206aa9d236b28fba579c1486ab530b0096f4f8a21033f911ee9dd03f8bac57a890412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000096a026d02047465737455030000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}},
}

var multiTest12faucet10point2k = MultiTest{
	Request: gen.MultiRequest{
		Outputs:      []*memo.Output{&test_tx.SetNameOutput},
		FaucetGetter: gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo10070}}, test_tx.Address1pkHash),
		FaucetSaver:  test_tx.GetFaucetSaverWithKey(test_tx.Key1String),
		Change:       wallet.GetChange(test_tx.Address2),
		KeyRing:      wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key2String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "a4cce9001bdb6aad34720082eb63b45331a50c07a4e07e4437a914881e3d7c39",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a4730440220141627e2a52d89cdfc6e77d307fb43bb2283b165085e76bb91ed4ce2991f910202206ad0d4b04d5bba096a715b925d96df6c06d42fd788c77564b3191d4bc96e0d86412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff02c9120000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588acab130000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}, {
		TxHash: "6b5b21dc879133bd33a26e0ad09073cb3fb3c385f92007b24d7992dc19fbc4bc",
		TxRaw:  "0100000001397c3d1e8814a937447ee0a4070ca53153b463eb82007234ad6adb1b00e9cca4000000006b483045022100b8ff5dc16968175189b2d4e16ebc6953e838871a44a2029f33fedb827b6f470d0220368f1ebfce4f229ea2d7c9adb85c3d56503fd832963fd15ee83fec926a677c57412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000096a026d010474657374f7110000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}},
}

var multiTests = []MultiTest{
	multiTest0empty,
	multiTest1like,
	multiTest2faucetLike,
	multiTest3faucetEmpty,
	multiTest4noFaucetChange,
	multiTest5faucetNotEnoughValue,
	multiTest6maxSend,
	multiTest7sendTooMuchWithToken,
	multiTest8faucetAndTokensLike,
	multiTest9faucetNotEnoughValue,
	multiTest10faucetNewPost,
	multiTest11maxSendWithToken,
	multiTest12faucet10point2k,
}
