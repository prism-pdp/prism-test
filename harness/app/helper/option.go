package helper

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pborman/getopt/v2"
)

var (
	HelpFlag *bool
	SimFlag *bool
	OptServer *string
	OptContractAddr *string
	OptSenderAddr *string
	OptSenderPrivKey *string
	OptPathCacheDir *string
	OptPathLogFile *string
	OptDetectedList *string
)

var (
	Server string
	ContractAddr string
	SenderAddr common.Address
	SenderPrivKey string
)

func SetupOpt() {
	data_dir := os.Getenv("PRISM_HARNESS_DATA_DIR")
	if data_dir == "" {
		data_dir = "/var/lib/prism-harness"
	}

	HelpFlag = getopt.BoolLong("help", 'h', "display help")
	SimFlag  = getopt.BoolLong("sim", 0, "simulation mode (disable blockchain)")
	OptServer = getopt.StringLong("server", 0, "", "server's URL")
	OptContractAddr = getopt.StringLong("contract", 0, "", "contract address")
	OptSenderAddr = getopt.StringLong("sender-addr", 0, "", "sender's address")
	OptSenderPrivKey = getopt.StringLong("sender-key", 0, "", "sender's private key")
	OptPathCacheDir = getopt.StringLong("cache", 0, data_dir + "/cache", "cache dir path")
	OptPathLogFile = getopt.StringLong("log", 0, data_dir + "/cache/prism.log", "log file path")
	OptDetectedList = getopt.StringLong("detected-list", 0, "", "detected list")
}
