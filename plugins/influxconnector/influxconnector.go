package influxconnector

import (
	"fmt"
	"net/url"
	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/theovassiliou/soundtouch-golang"
	"golang.org/x/exp/slices"
)

var name = "InfluxConnector"

const description = "Writes event data to InfluxDB "

const sampleConfig = `
# [influxDB]
## speakers for which messages should be logged. If empty, all 
# speakers = ["Office", "Kitchen"]

## log_messages describes the message types to be logged
## one or more of "ConnectionStateUpdated", "NowPlaying", "Volume"
## all if empty
# log_messages = ["ConnectionStateUpdated", "NowPlaying", "Volume"] 

## URL of the InfluxDB
# influxURL = "http://influxdb:8086"

## Database where to store the events
# database = "soundtouch"
#
## dry_run indicates that the plugin dumps lineporoto for the influxDB conncetion
## as curl statement. 
# dry_run = true
## 
`

// Config contains the configuration of the plugin
// Speakers list of SpeakerNames the handler is added. All if empty
// IgnoreMessages a list of message types to be ignored
type Config struct {
	InfluxURL   string   `toml:"influxURL"`
	Database    string   `toml:"database"`
	Speakers    []string `toml:"speakers"`
	LogMessages []string `toml:"log_messages"`
	DryRun      bool     `toml:"dry_run"`
}

// InfluxDB describes the plugin. It has a
// Config to store the configuration
// Plugin the plugin function
// suspended indicates that the plugin is temporarely suspended
type InfluxDB struct {
	Config
	Plugin    soundtouch.PluginFunc
	suspended bool
	noOfFails int
}

var influxDB = soundtouch.InfluxDB{
	BaseHTTPURL: url.URL{
		Scheme: "http",
		Host:   "localhost:8086",
	},
	Database: "soundtouch",
}

const maxNoOfFails = 20

// NewLogger creates a new Logger plugin with the configuration
func NewLogger(config Config) (d *InfluxDB) {
	d = &InfluxDB{}
	d.Config = config

	mLogger := log.WithFields(log.Fields{
		"Plugin": name,
	})
	if config.InfluxURL == "" {
		d.suspended = true
		return d
	}

	v, err := url.Parse(config.InfluxURL)
	if err != nil {
		mLogger.Infof("Not a valid URL: %v", config.InfluxURL)
		mLogger.Infof("Suspending plugin")
		d.suspended = true
		return d
	}

	influxDB.BaseHTTPURL = *v
	influxDB.Database = config.Database

	mLogger.Debugf("Initialised\n")
	return d
}

// Name returns the plugin name
func (d *InfluxDB) Name() string {
	return name
}

// Description returns a string explaining the purpose of this plugin
func (d *InfluxDB) Description() string { return description }

// SampleConfig returns text explaining how plugin should be configured
func (d *InfluxDB) SampleConfig() string { return sampleConfig }

// Terminate indicates that no further plugin will be executed on this speaker
func (d *InfluxDB) Terminate() bool { return false }

// Disable temporarely the execution of the plugin
func (d *InfluxDB) Disable() { d.suspended = true }

// Enable temporarely the execution of the plugin
func (d *InfluxDB) Enable() { d.suspended = false }

// IsEnabled returns true if the plugin is not suspened
func (d *InfluxDB) IsEnabled() bool { return !d.suspended }

// Execute runs the plugin with the given parameter
func (d *InfluxDB) Execute(pluginName string, update soundtouch.Update, speaker soundtouch.Speaker) {
	if d.suspended {
		return
	}

	mLogger := log.WithFields(log.Fields{
		"Plugin":        name,
		"Speaker":       speaker.Name(),
		"UpdateMsgType": reflect.TypeOf(update.Value).Name(),
	})
	mLogger.Debugln("Executing", pluginName)

	if len(d.LogMessages) > 0 && !slices.Contains(d.LogMessages, reflect.TypeOf(update.Value).Name()) {
		return
	}
	if len(d.Speakers) > 0 && !slices.Contains(d.Speakers, speaker.Name()) {
		return
	}

	v, _ := update.Lineproto(influxDB, &update)

	if !(d.Config.DryRun) && v != "" {
		result, err := influxDB.SetData("write", []byte(v))
		if err != nil {
			if d.noOfFails >= maxNoOfFails {
				d.suspended = true
				mLogger.Errorf("Failed %v times to connect. Disabling plugin.", d.noOfFails)
			} else {
				d.noOfFails = d.noOfFails + 1
				mLogger.Errorf("failed. No of fails %v", d.noOfFails)
				return
			}
		}
		d.noOfFails = 0
		mLogger.Debugf("succeeded: %v", string(result))

	} else if v != "" {
		fmt.Printf("curl -i -XPOST \"%v\" --data-binary '%v'\n", influxDB.WriteURL("write"), v)
	}
}
