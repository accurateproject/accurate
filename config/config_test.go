package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/DisposaBoy/JsonConfigReader"
	"github.com/accurateproject/accurate/utils"
)

func TestConfigCompareDefaults(t *testing.T) {
	data, err := ioutil.ReadFile("../data/conf/defaults/defaults.json")
	if err != nil {
		t.Error("error loading default file: ", err)
	}
	r := JsonConfigReader.New(bytes.NewBuffer(data))

	newConf := &Config{}
	defaultConfig.General.InstanceID = nil // don't use instance id
	if err := json.NewDecoder(r).Decode(newConf); err != nil {
		t.Error("error unmarshalling data: ", err)
	}
	if utils.ToJSON(defaultConfig) != utils.ToJSON(newConf) {
		t.Errorf("expected %s got %s", utils.ToJSON(defaultConfig), utils.ToJSON(newConf))
	}
	//t.Errorf("%+v", newConf.DiameterAgent.RequestProcessors[0].CcrFields[0].Value[0])
}

func TestConfigLoadSliceAppend(t *testing.T) {
	err := LoadPath("../data/conf/samples/dmtagent")
	if err != nil {
		t.Error("error loading path: ", err)
	}
	if len(defaultConfig.DiameterAgent.RequestProcessors) != 13 {
		log.Print("error appending request processors: ", utils.ToIJSON(defaultConfig.DiameterAgent.RequestProcessors), len((defaultConfig.DiameterAgent.RequestProcessors)))
	}
}

func TestEnvVariableReplace(t *testing.T) {
	Reset()
	os.Setenv("MONGO_PORT_27017_TCP_ADDR", "192.168.1.101")
	os.Setenv("MONGO_PORT_27017_TCP_PORT", "200")
	err := LoadBytes([]byte(`{
"data_db": {								// database used to store offline tariff plans and CDRs
	"db_type": "mongo",						// stor database type to use: <mysql|postgres>
	"db_host": "$$MONGO_PORT_27017_TCP_ADDR",					// the host to connect to
	"db_port": "$$MONGO_PORT_27017_TCP_PORT",						// the port to reach the stordb
  "db_name": "datadb",
  "db_user": "",
  "db_password": "",
},
}`), true)
	if err != nil {
		t.Error("error loading path: ", err)
	}
	if *defaultConfig.DataDb.Host != "192.168.1.101" ||
		*defaultConfig.DataDb.Port != "200" {
		t.Error("error replacing env variables: ", utils.ToIJSON(defaultConfig.TariffPlanDb))
	}

}
