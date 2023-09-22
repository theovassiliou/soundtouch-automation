package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/toml"
	"github.com/jpillora/opts"
	log "github.com/sirupsen/logrus"

	"github.com/theovassiliou/soundtouch-automation/plugins/autooff"
	"github.com/theovassiliou/soundtouch-automation/plugins/auxjoin"
	"github.com/theovassiliou/soundtouch-automation/plugins/episodecollector"
	"github.com/theovassiliou/soundtouch-automation/plugins/influxconnector"
	"github.com/theovassiliou/soundtouch-automation/plugins/logger"
	"github.com/theovassiliou/soundtouch-automation/plugins/magiczone"
	"github.com/theovassiliou/soundtouch-automation/plugins/telegram"
	"github.com/theovassiliou/soundtouch-automation/plugins/volumebutler"
	"github.com/theovassiliou/soundtouch-golang"
)

var conf = config{}

const header = `
# Sountouch Automation Configuration
# Created by %s
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

# Will be ignored if static_speakers defined
no_of_soundtouch_systems=7

# static_speakers=[
#     "192.168.178.21", 
#     "192.168.178.32", 
# ]
`

type config struct {
	global
	LogLevel     log.Level `help:"Log level, one of panic, fatal, error, warn or warning, info, debug, trace"`
	SampleConfig bool      `opts:"group=Configuration" help:"If set creates a sample config file that can be used later"`
	PidFile      []string  `opts:"group=Configuration" help:"Write a PID file, if set"`
	Config       string    `opts:"group=Soundtouch" help:"configuration file to load"`
}

type global struct {
	Interface             string   `opts:"group=Soundtouch" help:"network interface to listen"`
	NoOfSoundtouchSystems int      `opts:"group=Soundtouch" help:"Number of Soundtouch systems to scan for."`
	StaticSpeakers        []string `opts:"group=Soundtouch" help:"A static list of IPs of speakers to handle. Superseeds NoOfSoundtouchSystems if set."`
}
type tomlConfig struct {
	Global           global
	Logger           *logger.Config           `toml:"logger"`
	EpisodeCollector *episodecollector.Config `toml:"episodeCollector"`
	MagicZone        *magiczone.Config        `toml:"magicZone"`
	InfluxDB         *influxconnector.Config  `toml:"influxDB"`
	VolumeButler     *volumebutler.Config     `toml:"volumeButler"`
	AutoOff          *autooff.Config          `toml:"autoOff"`
	Telegram         *telegram.Config         `toml:"telegram"`
	AuxJoin          *auxjoin.Config          `toml:"auxjoin"`
}

func main() {
	defer tearDown()

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
		Version(FormatFullVersion("masteringsoundtouch", version, branch, commit, build)).
		Parse()

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
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
	buf, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := toml.Unmarshal(buf, &tConfig); err != nil {
		panic(err)
	}

	conf.global.NoOfSoundtouchSystems = tConfig.Global.NoOfSoundtouchSystems
	conf.global.Interface = tConfig.Global.Interface
	conf.global.StaticSpeakers = tConfig.Global.StaticSpeakers

	pl := initPlugins(tConfig, false)

	nConf := soundtouch.NetworkConfig{
		InterfaceName:     conf.global.Interface,
		NoOfSystems:       conf.global.NoOfSoundtouchSystems,
		StaticIPAddresses: conf.global.StaticSpeakers,
		Plugins:           pl,
	}

	if conf.PidFile != nil {
		createPIDFile(conf.PidFile)
	}

	// SearchDevices does not closes the channel
	speakerCh := soundtouch.SearchDevices(nConf)
	for speaker := range speakerCh {
		log.Infof("Found device %s-%s with IP %s\n", speaker.Name(), speaker.DeviceID(), speaker.IP)
	}

}

func printSampleConfig(pl []soundtouch.Plugin) bool {
	var sampleConfig strings.Builder

	fHeader := fmt.Sprintf(header, FormatFullVersion("masteringsoundtouch", version, branch, commit, build))
	sampleConfig.WriteString(fHeader)

	for _, aPlugin := range pl {
		sampleConfig.WriteString(aPlugin.SampleConfig())
	}

	fmt.Println(sampleConfig.String())

	return true
}

func createPIDFile(pidFile []string) {
	if len(pidFile) < 1 {
		return
	}

	f, err := os.Create(pidFile[0])
	if err != nil {
		log.Errorln(err)
		return
	}
	_, err = f.WriteString(strconv.Itoa(os.Getpid()))
	if err != nil {
		log.Errorln(err)
		f.Close()
		return
	}
	err = f.Close()
	if err != nil {
		log.Errorln(err)
		return
	}
	log.Debugf("PID-file %s successfully created", pidFile[0])
}

func initPlugins(tConfig tomlConfig, mock bool) []soundtouch.Plugin {
	pl := []soundtouch.Plugin{}

	if tConfig.Logger != nil {
		pl = append(pl, logger.NewLogger(*tConfig.Logger))
	}

	if tConfig.EpisodeCollector != nil {
		pl = append(pl, episodecollector.NewCollector(*tConfig.EpisodeCollector))
	}
	if tConfig.MagicZone != nil {
		pl = append(pl, magiczone.NewCollector(*tConfig.MagicZone))
	}

	if tConfig.InfluxDB != nil {
		pl = append(pl, influxconnector.NewLogger(*tConfig.InfluxDB))
	}

	if tConfig.VolumeButler != nil {
		pl = append(pl, volumebutler.NewVolumeButler(*tConfig.VolumeButler))
	}

	if tConfig.AutoOff != nil {
		pl = append(pl, autooff.NewObserver(*tConfig.AutoOff))
	}

	if tConfig.Telegram != nil {
		pl = append(pl, telegram.NewTelegramLogger(*tConfig.Telegram))
	}

	if tConfig.AuxJoin != nil {
		pl = append(pl, auxjoin.NewAuxJoin(*tConfig.AuxJoin))
	}

	return pl
}

func tearDown() {
	if len(conf.PidFile) < 1 {
		log.Debugln("No PID file to delete")
		return
	}

	err := os.Remove(conf.PidFile[0])

	if err != nil {
		log.Errorln(err)
		return
	}
	log.Debugf("File %s successfully deleted", conf.PidFile[0])
}
