package admin

/*type NetworkTxRequest struct {
	Hash     string
	HashByte []byte
}

func (r *NetworkTxRequest) Parse(reader io.Reader) error {
	if err := json.NewDecoder(reader).Decode(r); err != nil {
		return jerr.Get("error unmarshalling network tx request", err)
	}
	hash, err := chainhash.NewHashFromStr(r.Hash)
	if err != nil {
		return jerr.Get("error parsing tx hash for network request", err)
	}
	r.HashByte = hash.CloneBytes()
	return nil
}*/

type NetworkTxResponse struct {
	Tx Tx
}

type Tx struct {
	Hash    string
	Raw     string
	Inputs  []Input
	Outputs []Output
}

type Input struct {
}

type Output struct {
}
