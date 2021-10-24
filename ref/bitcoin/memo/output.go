package memo

import (
	"github.com/jchavannes/jgo/jerr"
)

type Output struct {
	Amount int64
	Script Script
}

func (o Output) GetType() OutputType {
	if o.Script != nil {
		return o.Script.Type()
	}
	return OutputTypeUnknown
}

func (o Output) GetValuePlusFee() (int64, error) {
	if o.Script != nil {
		outputSize, err := GetOutputSize(o.Script)
		if err != nil {
			return 0, jerr.Get("error getting script", err)
		}
		return outputSize + o.Amount, nil
	}
	return -1, jerr.New("error getting memo output fee, script not set")
}

func (o Output) GetPkScript() ([]byte, error) {
	if o.Script == nil {
		return nil, jerr.New("error script not set")
	}
	outputScript, err := o.Script.Get()
	if err != nil {
		return nil, jerr.Get("error creating output", err)
	}
	return outputScript, nil
}

func GetOutputSize(script Script) (int64, error) {
	pkScript, err := script.Get()
	if err != nil {
		return 0, jerr.Get("error getting script", err)
	}
	var scriptLen = int64(len(pkScript))
	var scriptLenBytes int64
	if scriptLen < 0xFD {
		scriptLenBytes = 1
	} else if scriptLen < 0xFFFF {
		scriptLenBytes = 3
	} else if scriptLen < 0xFFFFFFFF {
		scriptLenBytes = 5
	} else {
		scriptLenBytes = 9
	}
	return OutputValueSize + scriptLenBytes + scriptLen, nil
}
