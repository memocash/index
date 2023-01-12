package wallet

import (
	"fmt"
	"github.com/jchavannes/btcd/txscript"
	"strings"
)

type OpCode struct {
	Code byte
	Data []byte
}

func (o OpCode) IsData() bool {
	return IsDataOpCode(o.Code)
}

func (o OpCode) String() string {
	return txscript.GetOpCodeString(o.Code)
}

func IsDataOpCode(code byte) bool {
	return code >= txscript.OP_DATA_1 && code <= txscript.OP_PUSHDATA4
}

type OpCodes []OpCode

func (o OpCodes) String() string {
	var parts []string
	for _, opCode := range o {
		parts = append(parts, opCode.String())
		if len(opCode.Data) > 0 {
			parts = append(parts, fmt.Sprintf("%x", opCode.Data))
		}
	}
	return strings.Join(parts, " ")
}

func Decompile(script []byte) []OpCode {
	var opCodes []OpCode
	for i := 0; i < len(script); i++ {
		var opCode = OpCode{
			Code: script[i],
		}
		if IsDataOpCode(script[i]) {
			var dataLength, offset int
			switch script[i] {
			case txscript.OP_PUSHDATA4:
				offset = 4
				if i+offset >= len(script) {
					return []OpCode{}
				}
				dataLength = int(script[i+1]) + int(script[i+2])*256 + int(script[i+3])*256*256 + int(script[i+4])*256*256*256
			case txscript.OP_PUSHDATA2:
				offset = 2
				if i+offset >= len(script) {
					return []OpCode{}
				}
				dataLength = int(script[i+1]) + int(script[i+2])*256
			case txscript.OP_PUSHDATA1:
				offset = 1
				if i+offset >= len(script) {
					return []OpCode{}
				}
				dataLength = int(script[i+1])
			default:
				dataLength = int(script[i])
			}
			if i+offset+1+dataLength > len(script) {
				return []OpCode{}
			}
			opCode.Data = script[i+offset+1 : i+offset+1+dataLength]
			i += dataLength + offset
		}
		opCodes = append(opCodes, opCode)
	}
	return opCodes
}
