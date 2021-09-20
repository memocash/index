package config

import (
	"github.com/jchavannes/jgo/db_util"
	"math"
	"strconv"
)

const (
	ShardSingle = math.MaxUint32
)

type Shard struct {
	Min   uint32 `mapstructure:"MIN"`
	Max   uint32 `mapstructure:"MAX"`
	Total uint32 `mapstructure:"TOTAL"`
	Host  string `mapstructure:"HOST"`
	Port  int    `mapstructure:"PORT"`
}

func (s Shard) String() string {
	return db_util.GetShardString(uint(s.Min), uint(s.Total))
}

func (s Shard) GetHost() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

func GetShardConfig(shard uint32, configs []Shard) Shard {
	if len(configs) > 1 && configs[0].Total > 0 {
		shard = shard % configs[0].Total
	}
	for _, config := range configs {
		if config.Min <= shard && shard <= config.Max {
			return config
		}
	}
	return configs[0]
}
