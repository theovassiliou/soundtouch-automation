package logger

import (
	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/theovassiliou/soundtouch-golang"
)

var name = "Logger"

const sampleConfig = `
## Enabling logger plugin
# [logger]

## speakers for which messages should be logged. If empty, all 
# speakers = ["Office", "Kitchen"]

## ignore_messages describes the message types to be ignored
## one or more of "ConnectionStateUpdated", "NowPlaying", "Volume"
## all if empty
# ignore_messages = ["ConnectionStateUpdated"] 

`

const description = "Logs update messages"

// Config contains the configuration of the plugin
// Speakers list of SpeakerNames the handler is added. All if empty
// IgnoreMessages a list of message types to be ignored
type Config struct {
	Speakers       []string `toml:"speakers"`
	IgnoreMessages []string `toml:"ignore_messages"`
}

// Logger describes the plugin. It has a
// Config to store the configuration
// Plugin the plugin function
// suspended indicates that the plugin is temporarely suspended
type Logger struct {
	Config
	Plugin    soundtouch.PluginFunc
	suspended bool
}

// NewLogger creates a new Logger plugin with the configuration
func NewLogger(config Config) (d *Logger) {
	d = &Logger{}
	d.Config = config

	mLogger := log.WithFields(log.Fields{
		"Plugin": name,
	})

	mLogger.Debugf("Initialised\n")

	return d
}

// Name returns the plugin name
func (d *Logger) Name() string {
	return name
}

// Description returns a string explaining the purpose of this plugin
func (d *Logger) Description() string { return description }

// SampleConfig returns text explaining how plugin should be configured
func (d *Logger) SampleConfig() string { return sampleConfig }

// Terminate indicates that no further plugin will be executed on this speaker
func (d *Logger) Terminate() bool { return false }

// Disable temporarely the execution of the plugin
func (d *Logger) Disable() { d.suspended = true }

// Enable temporarely the execution of the plugin
func (d *Logger) Enable() { d.suspended = false }

// IsEnabled returns true if the plugin is not suspened
func (d *Logger) IsEnabled() bool { return !d.suspended }

// Execute runs the plugin with the given parameter
func (d *Logger) Execute(pluginName string, update soundtouch.Update, speaker soundtouch.Speaker) {
	if len(d.IgnoreMessages) > 0 && sliceContains(reflect.TypeOf(update.Value).Name(), d.IgnoreMessages) {
		return
	}
	if len(d.Speakers) > 0 && !sliceContains(speaker.Name(), d.Speakers) {
		return
	}

	mLogger := log.WithFields(log.Fields{
		"Plugin":        name,
		"Speaker":       speaker.Name(),
		"UpdateMsgType": reflect.TypeOf(update.Value).Name(),
	})
	mLogger.Debugln("Executing", pluginName)
	mLogger.Infof("%v\n", update)
}

func sliceContains(name string, list []string) bool {
	for _, s := range list {
		if name == s {
			return true
		}
	}
	return false
}
