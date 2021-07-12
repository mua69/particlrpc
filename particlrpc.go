package particlrpc

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type ParticlRpc struct {
	dataDir    string
	rpcHost    string
	rpcPort    int
	rpcAuth    string
	httpClient *http.Client
}

type Sat int64

type NetworkInfo struct {
	Version     int    `json:"version"`
	Subversion  string `json:"subversion"`
	Connections int    `json:"connections"`
}

type BlockchainInfo struct {
	Blocks int `json:"blocks"`
}

type StakingInfo struct {
	Staking                 bool    `json:"staking"`
	Errors                  string  `json:"errors"`
	Weight                  int64   `json:"weight"`
	Percentyearreward       float64 `json:"percentyearreward"`
	Moneysupply             float64 `json:"moneysupply"`
	Treasurydonationpercent float64 `json:"treasurydonationpercent"`
	Netstakeweight          int64   `json:"netstakeweight"`
	Expectedtime            int64   `json:"expectedtime"`
}

type BlockRewardKernelScript struct {
	Spendaddr string `json:"spendaddr"`
}

type BlockRewardOutputScript struct {
	Hex       string `json:"hex"`
	Spendaddr string `json:"spendaddr"`
}

type BlockRewardOutput struct {
	Script BlockRewardOutputScript `json:"script"`
	Value  float64                 `json:"value"`
}

type BlockReward struct {
	Blockhash    string                  `json:"blockhash"`
	Coinstake    string                  `json:"coinstake"`
	Stakereward  float64                 `json:"stakereward"`
	Blockreward  float64                 `json:"blockreward"`
	Kernelscript BlockRewardKernelScript `json:"kernelscript"`
	Outputs      []BlockRewardOutput     `json:"outputs"`
}

type ColdStakeUnspent struct {
	Height    int    `json:"height"`
	Value     Sat    `json:"value"`
	Addrspend string `json:"addrspend"`
}

type Block struct {
	Hash   string `json:"hash"`
	Time   int64  `json:"time"`
	Height int    `json:"height"`
}

type AddressDelta struct {
	Satoshis Sat    `json:"satoshis"`
	Txid     string `json:"txid"`
}

type TxVin struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

type ScriptPubKey struct {
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

type TxVout struct {
	Type         string       `json:"type"`
	ValueSat     Sat          `json:"valueSat"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type Tx struct {
	Vin       []TxVin  `json:"vin"`
	Vout      []TxVout `json:"vout"`
	Time      int64    `json:"time"`
	Blockhash string   `json:"blockhash"`
}

type Stakingoptions struct {
	Rewardaddress     string  `json:"rewardaddress"`
	Enabled           bool    `json:"enabled"`
	Time              int64   `json:"time"`
	Smsgfeeratetarget float64 `json:"smsgfeeratetarget"`
}

type SetStakingoptions struct {
	Rewardaddress     string  `json:"rewardaddress"`
	Enabled           bool    `json:"enabled"`
	Smsgfeeratetarget float64 `json:"smsgfeeratetarget"`
}

type rpcResponse struct {
	Result interface{}
	Err    string `json:"error"`
	Id     int
}

// NewParticlRpc creates a new ParticlRpc instance with default settings:
// dataDir: ".", rpcHost: "localhost", rpcPort: 51735
func NewParticlRpc() *ParticlRpc {
	rpc := new(ParticlRpc)

	rpc.dataDir = "."
	rpc.rpcHost = "localhost"
	rpc.rpcPort = 51735
	rpc.httpClient = &http.Client{}

	return rpc
}

// SetDataDirectoy sets the particld data directory. An emptry string is interepreted as "."
func (rpc *ParticlRpc) SetDataDirectoy(dir string) {
	if dir != "" {
		rpc.dataDir = dir
	} else {
		rpc.dataDir = "."
	}
}

// SetRpcHost sets the host to which the RPC call will connect. An empty string is interpreted as "localhost"
func (rpc *ParticlRpc) SetRpcHost(host string) {
	if host != "" {
		rpc.rpcHost = host
	} else {
		rpc.rpcHost = "localhost"
	}
}

// SetRpcPort sets the port to which the RPC call will connect. Value <= 0 will be interpreted as 51735.
func (rpc *ParticlRpc) SetRpcPort(port int) {
	if port > 0 {
		rpc.rpcPort = port
	} else {
		rpc.rpcPort = 51735
	}
}

// ReadConfig reads a JSON config file defining data dir (data_dir), rpc host ("rpc_host") and
// rpc port ("rpc_port").
func (rpc *ParticlRpc) ReadConfig(filename string) error {
	var cfg struct {
		DataDir string `json:"data_dir"`
		RpcHost string `json:"rpc_host"`
		RpcPort int    `json:"rpc_port"`
	}

	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return errors.Wrap(err, "Failed to read ParticlRpc config file")
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return errors.Wrapf(err, "Syntax error in ParticlRpc config file %s", filename)
	}

	if cfg.DataDir != "" {
		rpc.dataDir = cfg.DataDir
	}

	if cfg.RpcHost != "" {
		rpc.rpcHost = cfg.RpcHost
	}

	if cfg.RpcPort > 0 {
		rpc.rpcPort = cfg.RpcPort
	}

	return nil
}

// ReadPartRpcCookie reads an rpc authorization .cookie file from the data directory.
func (rpc *ParticlRpc) ReadPartRpcCookie() error {
	path := fmt.Sprintf("%s/.cookie", rpc.dataDir)
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return errors.Wrap(err, "Failed to read particld cookie file")
	}

	rpc.rpcAuth = strings.TrimSpace(string(data))

	return nil
}

// CallRpc executes rpc command <cmd> with arguments <args> at the particl daemon. Returned data is
// written to <res>, which must be a pointer to a data structure matching the command. If rpc command
// is wallet specific, the wallet name can be passed in <wallet>, otherwise an empty string must be passed.
func (rpc *ParticlRpc) CallRpc(cmd string, wallet string, args []interface{}, res interface{}) error {
	data, err := json.Marshal(map[string]interface{}{
		"method": cmd,
		"id":     2,
		"params": args,
	})

	if err != nil {
		return errors.Wrap(err, "JSON Marshal failed")
	}

	url := fmt.Sprintf("http://%s@%s:%d", rpc.rpcAuth, rpc.rpcHost, rpc.rpcPort)
	if wallet != "" {
		url += "/wallet/" + wallet
	}
	resp, err := rpc.httpClient.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return errors.Wrap(err, "Post failed")
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("Bad response status: %s", resp.Status)
	}

	response := rpcResponse{}
	response.Result = res

	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&response)
	if err != nil {
		return errors.Wrap(err, "partRpc: Decode JSON")
	}

	if response.Err != "" {
		return errors.Errorf("RPC response error: %s", response.Err)
	}

	return nil
}

//GetNetworkInfo executes the "getnetworkinfo" command and returns results.
func (rpc *ParticlRpc) GetNetworkInfo() (*NetworkInfo, error) {
	var args []interface{}
	var res NetworkInfo

	err := rpc.CallRpc("getnetworkinfo", "", args, &res)

	if err != nil {
		return nil, errors.Wrap(err, "ParticlRpc: getnetworkinfo failed")
	}

	return &res, nil
}

//GetBlockchainInfo executes the "getnetworkinfo" command and returns results.
func (rpc *ParticlRpc) GetBlockchainInfo() (*BlockchainInfo, error) {
	var args []interface{}
	var res BlockchainInfo

	err := rpc.CallRpc("getblockchaininfo", "", args, &res)

	if err != nil {
		return nil, errors.Wrap(err, "ParticlRpc: getblockchaininfo failed")
	}

	return &res, nil
}

//GetStakingInfo executes the "getnetworkinfo" command and returns results.
func (rpc *ParticlRpc) GetStakingInfo(wallet string) (*StakingInfo, error) {
	var args []interface{}
	var res StakingInfo

	err := rpc.CallRpc("getstakinginfo", wallet, args, &res)

	if err != nil {
		return nil, errors.Wrap(err, "ParticlRpc: getstakinginfo failed")
	}

	return &res, nil
}

//GetStakingInfo executes the "getnetworkinfo" command and returns results.
func (rpc *ParticlRpc) GetUptime() (int64, error) {
	var args []interface{}
	var res int64

	err := rpc.CallRpc("uptime", "", args, &res)

	if err != nil {
		return 0, errors.Wrap(err, "ParticlRpc: uptime failed")
	}

	return res, nil
}

//SetStakingOptions sets staking options (staking enabled, reward address) for specified wallet and returns
//current settings.
func (rpc *ParticlRpc) SetStakingOptions(enabled bool, rewardaddress string, smsgfeeratetarget float64,
	wallet string) (*Stakingoptions, error) {
	var args []interface{}
	var res struct {
		Stakingoptions Stakingoptions `json:"stakingoptions"`
	}

	var options SetStakingoptions

	options.Enabled = enabled
	options.Rewardaddress = rewardaddress
	options.Smsgfeeratetarget = smsgfeeratetarget

	args = append(args, "stakingoptions", options)

	err := rpc.CallRpc("walletsettings", wallet, args, &res)

	if err != nil {
		return nil, errors.Wrap(err, "ParticlRpc: set walletsettings failed")
	}

	return &res.Stakingoptions, nil
}

//GetStakingOptions retuns currently set staking options
func (rpc *ParticlRpc) GetStakingOptions(wallet string) (*Stakingoptions, error) {
	var args []interface{}
	var res struct {
		Stakingoptions Stakingoptions `json:"stakingoptions"`
	}

	args = append(args, "stakingoptions")

	err := rpc.CallRpc("walletsettings", wallet, args, &res)

	if err != nil {
		return nil, errors.Wrap(err, "ParticlRpc: get walletsettings failed")
	}

	return &res.Stakingoptions, nil
}
