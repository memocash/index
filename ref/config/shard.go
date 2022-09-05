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
	Shard uint32 `mapstructure:"SHARD"`
	Total uint32 `mapstructure:"TOTAL"`
	Host  string `mapstructure:"HOST"`
	Port  int    `mapstructure:"PORT"`
}

func (s Shard) String() string {
	return db_util.GetShardString(uint(s.Shard), uint(s.Total))
}

func (s Shard) GetHost() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

func (s Shard) Int() int {
	return int(s.Shard)
}

func GetShardConfig(shard uint32, configs []Shard) Shard {
	if len(configs) > 1 && configs[0].Total > 0 {
		shard = shard % configs[0].Total
	}
	for _, config := range configs {
		if config.Shard == shard {
			return config
		}
	}
	return configs[0]
}
