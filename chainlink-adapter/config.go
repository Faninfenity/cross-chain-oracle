// config.go - 全局配置读取模块
// 放置位置: ~/cross-chain-project/config.go
// 各微服务通过 import 使用，启动时自动读取同目录下的 config.toml

package main
import (
    "fmt"
    

    "github.com/BurntSushi/toml"
)
// AppConfig 全局配置结构体
type AppConfig struct {
	Fabric    FabricConfig    `toml:"fabric"`
	Fisco     FiscoConfig     `toml:"fisco"`
	Chainlink ChainlinkConfig `toml:"chainlink"`
	IPFS      IPFSConfig      `toml:"ipfs"`
	Ports     PortsConfig     `toml:"ports"`
}

type FabricConfig struct {
	BasePath    string `toml:"base_path"`
	Channel     string `toml:"channel"`
	Chaincode   string `toml:"chaincode"`
	PeerAddress string `toml:"peer_address"`
	CliPath     string `toml:"cli_path"`
}

type FiscoConfig struct {
	ConsoleDir   string `toml:"console_dir"`
	ContractAddr    string `toml:"contract_addr"`
	ReputationAddr string `toml:"reputation_addr"`
	ContractName string `toml:"contract_name"`
}

type ChainlinkConfig struct {
	NodeURL  string `toml:"node_url"`
	JobID    string `toml:"job_id"`
	Email    string `toml:"email"`
	Password string `toml:"password"`
}

type IPFSConfig struct {
	ApiURL string `toml:"api_url"`
}

type PortsConfig struct {
	FabricAdapter string `toml:"fabric_adapter"`
	FiscoWriter   string `toml:"fisco_writer"`
	AutoTrigger   string `toml:"auto_trigger"`
	VerifierUI    string `toml:"verifier_ui"`
	IssuerUI      string `toml:"issuer_ui"`
}

// Cfg 全局配置实例，程序启动时加载
var Cfg AppConfig

// LoadConfig 加载配置文件
// 优先查找可执行文件同目录下的 config.toml
// 其次查找项目根目录
func LoadConfig() error {
    if _, err := toml.DecodeFile("config.toml", &Cfg); err != nil {
        return fmt.Errorf("配置文件解析失败: %v", err)
    }
    fmt.Println("[Config] 配置已加载: config.toml")
    return nil
}
// FabricEnv 返回 Fabric peer 命令所需的环境变量列表
func (f *FabricConfig) FabricEnv() []string {
	return []string{
		"FABRIC_CFG_PATH=" + f.BasePath + "/config",
		"CORE_PEER_TLS_ENABLED=true",
		"CORE_PEER_LOCALMSPID=Org1MSP",
		"CORE_PEER_TLS_ROOTCERT_FILE=" + f.BasePath + "/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt",
		"CORE_PEER_MSPCONFIGPATH=" + f.BasePath + "/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp",
		"CORE_PEER_ADDRESS=" + f.PeerAddress,
	}
}

// PeerBin 返回 peer 二进制文件路径
func (f *FabricConfig) PeerBin() string {
	return f.BasePath + "/bin/peer"
}

// WebhookURL 返回完整的 Chainlink Webhook URL
func (c *ChainlinkConfig) WebhookURL() string {
	return fmt.Sprintf("%s/v2/jobs/%s/runs", c.NodeURL, c.JobID)
}

// SessionURL 返回 Chainlink 登录接口 URL
func (c *ChainlinkConfig) SessionURL() string {
	return c.NodeURL + "/sessions"
}

// IPFSAddURL 返回 IPFS 上传接口 URL
func (i *IPFSConfig) AddURL() string {
	return i.ApiURL + "/api/v0/add"
}
