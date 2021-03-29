package testing

import "github.com/chenjianmei111/go-state-types/abi"

func MakePID(input string) abi.PeerID {
	return abi.PeerID([]byte(input))
}
