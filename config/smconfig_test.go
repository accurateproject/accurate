package config

import (
	"reflect"
	"testing"

	"github.com/accurateproject/accurate/utils"
)

func TesSmFsConfigLoadFromJsonCfg(t *testing.T) {
	smFsJsnCfg := &SmFsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Create_cdr:     utils.BoolPointer(true),
		Subscribe_park: utils.BoolPointer(true),
		Event_socket_conns: &[]*FsConnJsonCfg{
			&FsConnJsonCfg{
				Address:    utils.StringPointer("1.2.3.4:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
			&FsConnJsonCfg{
				Address:    utils.StringPointer("2.3.4.5:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
		},
	}
	eSmFsConfig := &SmFsConfig{Enabled: true,
		CreateCdr:     true,
		SubscribePark: true,
		EventSocketConns: []*FsConnConfig{
			&FsConnConfig{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 5},
			&FsConnConfig{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 5},
		},
	}
	smFsCfg := new(SmFsConfig)
	if err := smFsCfg.loadFromJsonCfg(smFsJsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSmFsConfig, smFsCfg) {
		t.Error("Received: ", smFsCfg)
	}
}
