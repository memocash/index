package gen_test

import (
	"fmt"
	"github.com/jchavannes/jgo/jlog"
	"github.com/memocash/server/ref/bitcoin/memo"
	"github.com/memocash/server/ref/bitcoin/tx/gen"
	"github.com/memocash/server/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/server/ref/bitcoin/wallet"
	"testing"
)

type MultiTest struct {
	Request  gen.MultiRequest
	Error    string
	TxHashes []test_tx.TxHash
}

func TestMulti(t *testing.T) {
	for j, multiTest := range multiTests {
		txs, err := gen.Multi(multiTest.Request)
		if testing.Verbose() {
			jlog.Logf("MultiTest %d:\n", j)
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
	Error:   gen.NotEnoughValueErrorText,
}

var multiTest1like = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{&test_tx.LikeEmptyPostOutput},
		Getter:  gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.Address1InputUtxo100k}}, test_tx.Address1pkHash),
		Change:  wallet.GetChange(test_tx.Address1),
		KeyRing: wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key1String)),
	},
	TxHashes: []test_tx.TxHash{{
		TxHash: "d4c01b19b50f249d04779bd7acc510026fb215bac9eaa61c4fe56f6a3693f8ca",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a4730440220203b4b23d15054ad92ecccb99f4d2d82b7da3982950b0b3a638dda6d3004847c02201bd7ba83905277b4995f0d6564dad35bd2edd583fe651d2be65d58efd55ba325412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2b2850100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
		TxHash: "96b9ea8dadd44e136fc5f9c11dabf4a86e733fdb542d63de513121b20596a9bb",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b48304502210098555225f062a1c59f4c9d07619b860caa9c3c7ae60ccb8574f663412bd6f124022067ea15d9ddf632d39365020da6d958733934d04f7bc03d7480ab58127e84cf45412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff022e260000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac905f0100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}, {
		TxHash: "f229b1680fdf9769453c4abd3b17601aff748566c59f6ee37adbd0e3c8938dc9",
		TxRaw:  "0100000001bba99605b2213151de632d54db3f736ea8f4ab1dc1f9c56f134ed4ad8deab996000000006b483045022100abd4c9ad235d002250290448cb1e4d4d9b681900cff99916633496935c6c24df02203d4ceb6981eeaa81dc8a99bb6c1f897fa4ace418a0bbfdccbea01cc54b663ade412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d240250000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}},
}

var multiTest3faucetEmpty = MultiTest{
	Request: gen.MultiRequest{
		Outputs: []*memo.Output{&test_tx.LikeEmptyPostOutput},
		Change:  wallet.GetChange(test_tx.Address2),
		KeyRing: wallet.GetSingleKeyRing(test_tx.GetPrivateKey(test_tx.Key2String)),
	},
	Error: gen.NotEnoughValueErrorText,
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
		TxHash: "8ce5d634222c8b0a8f60691e2743540d8022912a91662a0f512911c39140e0a4",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b48304502210082b1b4657911a9d6d41c19a2d0a4cab5826873ce42267493d48326c21c1adbd1022027a6a2c6b8716a9b7c0dea6634a2a2acb201ad16a45e36a0d429f68a32d20df9412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff01801e0000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
	}, {
		TxHash: "465cea0071fec521f937bc0fbbfb772e25d370f1459c032e53b2f17dbefd22e6",
		TxRaw:  "0100000001a4e04091c31129510f2a66912a9122800d5443271e69608f0a8b2c2234d6e58c000000006b483045022100cf8dc959074e1bce31237c3260713dd0dfd400fbf0333ae46df62062175bf70302205236db3bbdbf0ddad4cb68cefd127d9564d91414e6c59e2bf1f118ee8d3fea8b412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d2921d0000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
	Error: gen.BelowDustLimitErrorText,
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
		TxHash: "e6c88c318f8bc79780c8a49df3af5e0481684df74ee2b883f22a67f012883f9b",
		TxRaw:  "0100000002290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b4830450221009e9f8e1b950e521f64d71d0f288f47f9e63818099cb043961e78b3f4eacd713d022048456da91de357f4b4ea9e1fc227322158a72a61a0196a761c163390c242ac42412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4010000006a473044022060b76a89f140f40000019fffae019bba4ed88319600d79202de10464c02af40b022068fd959894a56028db0bd27ce91647ed0e3cc10a7e7234e8744318dce1b1f871412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff019e080000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
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
		TxHash: "e30d836dafea933be398e9247a3fb62921bd713bf4dc343c63da287fb59f0453",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100be7499ce179fb757f6a250e292c92c94d369d4c59d18bcfd9471938a8e0e150702202d95ae2690da22101ca9eb69b08aab01881803c17890a38baf090c8e5a51b707412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0110070000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
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
	Error: gen.NotEnoughValueErrorText,
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
		TxHash: "96b9ea8dadd44e136fc5f9c11dabf4a86e733fdb542d63de513121b20596a9bb",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b48304502210098555225f062a1c59f4c9d07619b860caa9c3c7ae60ccb8574f663412bd6f124022067ea15d9ddf632d39365020da6d958733934d04f7bc03d7480ab58127e84cf45412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff022e260000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac905f0100000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}, {
		TxHash: "f229b1680fdf9769453c4abd3b17601aff748566c59f6ee37adbd0e3c8938dc9",
		TxRaw:  "0100000001bba99605b2213151de632d54db3f736ea8f4ab1dc1f9c56f134ed4ad8deab996000000006b483045022100abd4c9ad235d002250290448cb1e4d4d9b681900cff99916633496935c6c24df02203d4ceb6981eeaa81dc8a99bb6c1f897fa4ace418a0bbfdccbea01cc54b663ade412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d240250000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
		TxHash: "4a741c24e5bf9cca08bd1f786d38acfb3a022bbc579805037b9277c0c2ef4380",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b4830450221008e2983621a8975373e7b48242a6ffc6f9c67185864dac234904e6648bee67fbc02207cacc5ab479fbd40fe8f48444f4471588c537a22843f1baa69d1ecb99a3f434c412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0128030000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
	}, {
		TxHash: "54ba8caca32fc6a3fb0d38c19e8d6b545ac4415c7c6677d45ad239c7ba787908",
		TxRaw:  "01000000018043efc2c077927b03059857bc2b023afbac386d781fbd08ca9cbfe5241c744a000000006a4730440220631ff1b58afb58f72eb1528545a0799bf8dc7d5ef2440c90e1629aac6d6ec90302203fc00d410f8a7c40ae754fbfe8af806121a08d450c8e94b58e9ae0284dfa96b0412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000256a026d042043ec7a579f5561a42a7e9637ad4156672735a658be2752181801f723ba3316d23a020000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
		TxHash: "91d9902c8600a86afe160ed32aa53a1bfedf83f58cb94599a8e2d253e07e76a3",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006b483045022100df4ce456352a9ab5d75612ca828c5e69a25097ebb9aed3689f9522ffe2d064ce02206612be039522d04d30bf6ae360d490c8670de96e1d41443424b2b4c2bc39160b412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff0127040000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588ac00000000",
	}, {
		TxHash: "55d223d0d2f4e16c742929ded20a1ab5a8f0408bc13e3759fa905917f93debb6",
		TxRaw:  "0100000001a3767ee053d2e2a89945b98cf583dffe1b3aa52ad30e16fe6aa800862c90d991000000006b483045022100c05631cfe9c5e8e0721ccc81ade28a91c8c134f5687e3cc1443ba7921ca5cc2f0220401e824361c930393ce20672f9bd47398c1745953f573b32c69daef29fc8da5f412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000096a026d02047465737455030000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
		TxHash: "1651d516f25226298ec62c4c60568ee6d8dec6b00b0caaed01b6fde3641defc5",
		TxRaw:  "0100000001290c9e545233529c68f1efac662cb3370df17d08cdbaa7e63e04284e670ffef4000000006a473044022037d10ec738d73ba732dceb2a505cd623e2786f65a7df016c9ab43a256738e77b02206d8575b4144f062b5753304cb8d7889ec0b2ab6424d9235d0ae92118f2865a51412103065e9c67d6ef37c1b08f88d74a4b2090aa8d69f2e6ab5c116f60f05a78f2ededffffffff02c9120000000000001976a9140d4cd6490ddf863bbdf5c34d8ef1aebfd45c210588acab130000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
	}, {
		TxHash: "76bfc611ab850392b754c290b45494d7c9fcdbe7911c9362ba5c6c791d4e95d6",
		TxRaw:  "0100000001c5ef1d64e3fdb601edaa0c0bb0c6ded8e68e56604c2cc68e292652f216d55116000000006a47304402203303776f7de37432659723d58215f3468037e029039badb0e9f45ffbc7f9c4c1022047c1bbad54d227ff94de5699353f187807f0f530a9dd8e63db57a8b7efe89679412102de3c9a32a16686498b8e71efa73902f679e977bf1f8381538faf3e68737f92cdffffffff020000000000000000096a026d010474657374f7110000000000001976a914fc393e225549da044ed2c0011fd6c8a799806b6288ac00000000",
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
