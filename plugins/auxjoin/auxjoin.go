package auxjoin

import (
	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/theovassiliou/soundtouch-golang"
	"golang.org/x/exp/slices"
)

var name = "AuxJoin"

const description = "If Aux is selected speaker joins existing stream"

const sampleConfig = `
## Enabling the auxjoin plugin
# [auxjoin]

## ordered list of speakers that can join in zones. All if empty.
# speakers = ["Office", "Kitchen", "Schlafzimmer", "Schrank"]

`

// Config contains the configuration of the plugin
// Speakers list of SpeakerNames the handler is added. All if empty
type Config struct {
	Speakers []string `toml:"-"`
}

// AuxJoin describes the plugin. It has a
// Config to store the configuration
// Plugin the plugin function
// suspended indicates that the plugin is temporarely suspended
type AuxJoin struct {
	Config
	Plugin    soundtouch.PluginFunc
	suspended bool
}

// NewAuxJoin creates a new Collector plugin with the configuration
func NewAuxJoin(config Config) (d *AuxJoin) {
	d = &AuxJoin{}
	d.Config = config

	mLogger := log.WithFields(log.Fields{
		"Plugin": name,
	})

	mLogger.Debugf("Initialised\n")

	return d
}

// Name returns the plugin name
func (d *AuxJoin) Name() string {
	return name
}

// Description returns a string explaining the purpose of this plugin
func (d *AuxJoin) Description() string { return description }

// SampleConfig returns text explaining how plugin should be configured
func (d *AuxJoin) SampleConfig() string { return sampleConfig }

// Terminate indicates that no further plugin will be executed on this speaker
func (d *AuxJoin) Terminate() bool { return false }

// Disable temporarely the execution of the plugin
func (d *AuxJoin) Disable() { d.suspended = true }

// Enable temporarely the execution of the plugin
func (d *AuxJoin) Enable() { d.suspended = false }

// IsEnabled returns true if the plugin is not suspened
func (d *AuxJoin) IsEnabled() bool { return !d.suspended }

// Execute runs the plugin with the given parameter
// AuxJoin adds the auxed speaker to an existing stream, according to the following rules
// 1. Look whether there is a zone playing
// 		If yes, add this speaker to the zone
//		If no, continue
// 2. Look wether the is speaker playing a compatible source
//  	If yes, take the first speaker
//  		Make the first speaker to master and auxed speaker as slave
//  	If no, exit

func (d *AuxJoin) Execute(pluginName string, update soundtouch.Update, speaker soundtouch.Speaker) {
	if !(update.Is("NowPlaying")) {
		return
	}

	if len(d.Speakers) > 0 && !slices.Contains(d.Speakers, speaker.Name()) {
		return
	}

	mLogger := log.WithFields(log.Fields{
		"Plugin":        name,
		"Speaker":       speaker.Name(),
		"UpdateMsgType": reflect.TypeOf(update.Value).Name(),
	})

	mLogger.Debugln("Executing", pluginName)

	np := update.Value.(soundtouch.NowPlaying)
	if !(np.PlayStatus == soundtouch.PlayState && np.Source == soundtouch.Aux) {
		mLogger.Traceln("PlayStatus != PlayState and not AUX--> Done!")
		return
	}
	mLogger.Traceln("Selected AUX")
	mLogger.Traceln("Searching for a master")

	for _, aKnownDevice := range soundtouch.GetKnownDevices() {
		if aKnownDevice.IsMaster() {
			newZone := soundtouch.NewZone(*aKnownDevice, speaker)
			aKnownDevice.AddZoneSlave(newZone)
			mLogger.Debugf("added %v to master %v\n", speaker.Name(), aKnownDevice.Name())
			return
		}
	}
	mLogger.Traceln("Haven't found a master. Continuing the search")

	for _, aKnownDevice := range soundtouch.GetKnownDevices() {
		np, _ := aKnownDevice.NowPlaying()
		if np.PlayStatus == soundtouch.PlayState &&
			(np.Source == soundtouch.LocalInternetRadio ||
				np.Source == soundtouch.StoredMusic ||
				np.Source == soundtouch.Spotify ||
				np.Source == soundtouch.Alexa) {
			mLogger.Debugf("Found a suitable source %v with %v\n", np.Source, aKnownDevice.Name())
			newZone := soundtouch.NewZone(*aKnownDevice, speaker)
			mLogger.Debugf("Creating new zone with %v as master.\n", newZone.Master)
			aKnownDevice.SetZone(newZone)
			return
		}
	}
	mLogger.Tracef("AUX on %v no effect as no target to join\n", speaker.Name())
}
