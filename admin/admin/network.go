package admin

/*type NetworkTxRequest struct {
	Hash     string
	HashByte []byte
}

func (r *NetworkTxRequest) Parse(reader io.Reader) error {
	if err := json.NewDecoder(reader).Decode(r); err != nil {
		return fmt.Errorf("error unmarshalling network tx request; %w", err)
	}
	hash, err := chainhash.NewHashFromStr(r.Hash)
	if err != nil {
		return fmt.Errorf("error parsing tx hash for network request; %w", err)
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
