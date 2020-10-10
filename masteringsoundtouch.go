package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/influxdata/toml"
	"github.com/jpillora/opts"
	log "github.com/sirupsen/logrus"

	"github.com/theovassiliou/soundtouch-golang"
	"github.com/theovassiliou/soundtouch-master/plugins/autooff"
	"github.com/theovassiliou/soundtouch-master/plugins/episodecollector"
	"github.com/theovassiliou/soundtouch-master/plugins/influxconnector"
	"github.com/theovassiliou/soundtouch-master/plugins/logger"
	"github.com/theovassiliou/soundtouch-master/plugins/magiczone"
	"github.com/theovassiliou/soundtouch-master/plugins/telegram"
	"github.com/theovassiliou/soundtouch-master/plugins/volumebutler"
)

var conf = config{}

const header = `
# Sountouch Automation Configuration
#
# Soundtouch Automation is entirely plugin driven. All functionality is perfomed by 
# plugins.
#
# Plugins must be declared in here to be active.
# To deactivate a plugin, comment out the name and any variables.
#
# Where parameters can be set in the config file, environemnt variables or via command-line flags
# the order of precedence is as follows: 
#     defaults < config file < environment variables < command-line flags. 

# Global section contains global, plugin independent parameters
[global]
interface="en0"
no_of_soundtouch_systems=7

`

type config struct {
	global
	LogLevel     log.Level `help:"Log level, one of panic, fatal, error, warn or warning, info, debug, trace"`
	SampleConfig bool      `opts:"group=Configuration" help:"If set creates a sample config file that can be used later"`
	Config       string    `opts:"group=Soundtouch" help:"configuration file to load"`
}

type global struct {
	Interface             string `opts:"group=Soundtouch" help:"network interface to listen"`
	NoOfSoundtouchSystems int    `opts:"group=Soundtouch" help:"Number of Soundtouch systems to scan for."`
}
type tomlConfig struct {
	Global           global
	Logger           logger.Config           `toml:"logger"`
	EpisodeCollector episodecollector.Config `toml:"episodeCollector"`
	MagicZone        magiczone.Config        `toml:"magicZone"`
	InfluxDB         influxconnector.Config  `toml:"influxDB"`
	VolumeButler     volumebutler.Config     `toml:"volumeButler"`
	AutoOff          autooff.Config          `toml:"autoOff"`
	Telegram         telegram.Config         `toml:"telegram"`
}

func main() {
	var tConfig tomlConfig
	conf = config{
		global: global{
			Interface:             "en0",
			NoOfSoundtouchSystems: -1,
		},
		SampleConfig: false,
		LogLevel:     log.InfoLevel,
		Config:       "config.toml",
	}

	//parse config
	opts.New(&conf).
		Parse()

	log.SetLevel(conf.LogLevel)

	if conf.SampleConfig {
		printSampleConfig(initPlugins(tConfig, true))
		log.Infoln("Dumped sample config file")
		os.Exit(0)
	}

	f, err := os.Open(conf.Config)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := toml.Unmarshal(buf, &tConfig); err != nil {
		panic(err)
	}
	conf.global.NoOfSoundtouchSystems = tConfig.Global.NoOfSoundtouchSystems

	pl := initPlugins(tConfig, false)

	nConf := soundtouch.NetworkConfig{
		InterfaceName: conf.global.Interface,
		NoOfSystems:   conf.global.NoOfSoundtouchSystems,
		Plugins:       pl,
	}

	// SearchDevices does not closes the channel
	speakerCh := soundtouch.SearchDevices(nConf)
	for speaker := range speakerCh {
		log.Infof("Found device %s-%s with IP %s\n", speaker.Name(), speaker.DeviceID(), speaker.IP)
	}

}

func printSampleConfig(pl []soundtouch.Plugin) bool {
	var sampleConfig strings.Builder

	sampleConfig.WriteString(header)

	for _, aPlugin := range pl {
		sampleConfig.WriteString(aPlugin.SampleConfig())
	}

	fmt.Println(sampleConfig.String())

	return true
}

func initPlugins(tConfig tomlConfig, mock bool) []soundtouch.Plugin {
	pl := []soundtouch.Plugin{}

	pl = append(pl, logger.NewLogger(tConfig.Logger))
	pl = append(pl, episodecollector.NewCollector(tConfig.EpisodeCollector))
	pl = append(pl, magiczone.NewCollector(tConfig.MagicZone))
	pl = append(pl, influxconnector.NewLogger(tConfig.InfluxDB))
	pl = append(pl, volumebutler.NewVolumeButler(tConfig.VolumeButler))
	pl = append(pl, autooff.NewObserver(tConfig.AutoOff))
	pl = append(pl, telegram.NewTelegramLogger(tConfig.Telegram))
	return pl
}
