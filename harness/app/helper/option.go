package helper

import (
	"github.com/pborman/getopt/v2"
)

var (
	HelpFlag *bool
	SimFlag *bool
	OptServer *string
	OptContractAddr *string
	OptSenderAddr *string
	OptSenderPrivKey *string
)

var (
	Server string
	ContractAddr string
	SenderAddr string
	SenderPrivKey string
)

func SetupOpt() {
	HelpFlag = getopt.BoolLong("help", 'h', "display help")
	SimFlag  = getopt.BoolLong("sim", 0, "simulation mode (disable blockchain)")
	OptServer = getopt.StringLong("server", 0, "", "server's URL")
	OptContractAddr = getopt.StringLong("contract", 0, "", "contract address")
	OptSenderAddr = getopt.StringLong("sender-addr", 0, "", "sender's address")
	OptSenderPrivKey = getopt.StringLong("sender-key", 0, "", "sender's private key")
}
