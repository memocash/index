package wallet

type Change struct {
	Main Address
	Slp  Address
}

func (c Change) GetSlp() Address {
	if c.Slp.IsSet() {
		return c.Slp
	}
	return c.Main
}

func GetChange(address Address) Change {
	return Change{Main: address}
}
