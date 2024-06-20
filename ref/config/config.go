package config

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

const (
	FlagConfig  = "config"
	FlagProfile = "profile"

	Localhost            = "127.0.0.1"
	DefaultAdminPort     = 26768
	DefaultBroadcastPort = 26769
	DefaultGraphQLPort   = 26770
	DefaultQueue0Port    = 26780
	DefaultQueue1Port    = 26781
	DefaultShard0Port    = 26790
	DefaultShard1Port    = 26791
	DefaultServerPort    = 19021

	DefaultInitBlock       = "00000000000000000038a66316b28503ca99d50f184b27cb2152d77ae6a38a12"
	DefaultInitBlockParent = "000000000000000000925634d697d3dcd7a8f5aef312f043f4cb278fd9152baa"
	DefaultInitBlockHeight = 625306
	DefaultBlocksToConfirm = 5

	DefaultDataDir = "db/data"
)

type Config struct {
	NodeHost string `mapstructure:"NODE_HOST"`

	InitBlock       string `mapstructure:"INIT_BLOCK"`
	InitBlockHeight uint   `mapstructure:"INIT_BLOCK_HEIGHT"`
	InitBlockParent string `mapstructure:"INIT_BLOCK_PARENT"`

	BlocksToConfirm uint `mapstructure:"BLOCKS_TO_CONFIRM"`

	ServerHost string `mapstructure:"SERVER_HOST"`
	ServerPort int    `mapstructure:"SERVER_PORT"`

	QueueShards []Shard `mapstructure:"QUEUE_SHARDS"`

	SaveMetrics bool `mapstructure:"SAVE_METRICS"`

	GraphQLPort   uint `mapstructure:"GRAPHQL_PORT"`
	AdminPort     uint `mapstructure:"ADMIN_PORT"`
	BroadcastPort int  `mapstructure:"BROADCAST_PORT"`

	DataDir string `mapstructure:"DATA_DIR"`

	DataPrefix             string `mapstructure:"DATA_PREFIX"`
	OpenFilesCacheCapacity int    `mapstructure:"OPEN_FILES_CACHE_CAPACITY"` // In MB
	CompactionDataSize     int    `mapstructure:"COMPACTION_DATA_SIZE"`

	ClusterShards []Shard `mapstructure:"CLUSTER_SHARDS"`

	ProcessLimit struct {
		Utxos int `mapstructure:"UTXOS"`
	} `mapstructure:"PROCESS_LIMIT"`

	Influx InfluxConfig `mapstructure:"INFLUX"`
}

var _config Config

var DefaultConfig = Config{
	NodeHost:        "[bitcoind]:8333",
	InitBlock:       DefaultInitBlock,
	InitBlockHeight: DefaultInitBlockHeight,
	InitBlockParent: DefaultInitBlockParent,
	BlocksToConfirm: DefaultBlocksToConfirm,
	ServerHost:      Localhost,
	ServerPort:      DefaultServerPort,
	AdminPort:       DefaultAdminPort,
	GraphQLPort:     DefaultGraphQLPort,
	BroadcastPort:   DefaultBroadcastPort,
	DataDir:         DefaultDataDir,
	QueueShards: []Shard{{
		Shard: 0,
		Total: 2,
		Host:  Localhost,
		Port:  DefaultQueue0Port,
	}, {
		Shard: 1,
		Total: 2,
		Host:  Localhost,
		Port:  DefaultQueue1Port,
	}},
	ClusterShards: []Shard{{
		Shard: 0,
		Total: 2,
		Host:  Localhost,
		Port:  DefaultShard0Port,
	}, {
		Shard: 1,
		Total: 2,
		Host:  Localhost,
		Port:  DefaultShard1Port,
	}},
}

func Init(cmd *cobra.Command) error {
	fmt.Print(GetHost(8333))
	config, _ := cmd.Flags().GetString(FlagConfig)
	if config != "" && !strings.HasPrefix(config, "config-") {
		config = "config-" + config
	} else if config == "" {
		config = "config"
	}
	viper.SetConfigName(config)
	viper.AddConfigPath("$HOME/.index")
	viper.AddConfigPath(".")
	viper.AddConfigPath(".config/index")
	if err := viper.ReadInConfig(); err != nil {
		// Config not found, use default
		_config = DefaultConfig
		return nil
	}
	if err := viper.Unmarshal(&_config); err != nil {
		return fmt.Errorf("error unmarshalling config; %w", err)
	}
	if len(_config.ClusterShards) != len(_config.QueueShards) {
		return fmt.Errorf("error config cluster shards and queue shards must be the same length")
	}
	return nil
}

func GetNodeHost() string {
	return _config.NodeHost
}

func GetInitBlock() string {
	return _config.InitBlock
}

func GetInitBlockHeight() uint {
	return _config.InitBlockHeight
}

func GetInitBlockParent() string {
	return _config.InitBlockParent
}

func GetBlocksToConfirm() uint {
	return _config.BlocksToConfirm
}

func GetQueueShards() []Shard {
	return _config.QueueShards
}

func GetClusterShards() []Shard {
	return _config.ClusterShards
}

func GetTotalShards() uint32 {
	if len(_config.QueueShards) == 0 {
		return 0
	}
	return _config.QueueShards[0].Total
}

func GetTotalClusterShards() uint32 {
	if len(_config.ClusterShards) == 0 {
		return 0
	}
	return _config.ClusterShards[0].Total
}

func GetSaveMetrics() bool {
	return _config.SaveMetrics
}

func GetServerPort() int {
	return _config.ServerPort
}

func GetProcessLimitUtxos() int {
	return _config.ProcessLimit.Utxos
}

func GetAdminPort() uint {
	return _config.AdminPort
}

func GetGraphQLPort() uint {
	return _config.GraphQLPort
}

func GetBroadcastRpc() RpcConfig {
	return RpcConfig{
		Host: Localhost,
		Port: _config.BroadcastPort,
	}
}

func GetDataPrefix() string {
	return _config.DataPrefix
}

func GetOpenFilesCacheCapacity() int {
	return _config.OpenFilesCacheCapacity
}

func GetCompactionDataSize() int {
	return _config.CompactionDataSize
}

func GetHost(port uint) string {
	fmt.Printf("[%s]:%d", Localhost, port)
	return fmt.Sprintf("[%s]:%d", Localhost, port)
}

func GetDataDir() string {
	return _config.DataDir
}

func GetInfluxConfig() InfluxConfig {
	return _config.Influx
}
