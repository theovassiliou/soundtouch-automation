package magiczone

import (
	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/theovassiliou/soundtouch-golang"
	"golang.org/x/exp/slices"
)

var name = "MagicZone"

const description = "Groups speaker that play the same content in a zone"

const sampleConfig = `
## Enabling the magicZone plugin
# [magicZone]

## ordered list of speakers that should be grouped in zones. All if empty.
# speakers = ["Office", "Kitchen"]

`

// Config contains the configuration of the plugin
// Speakers list of SpeakerNames the handler is added. All if empty
type Config struct {
	Speakers []string `toml:"-"`
}

// MagicZone describes the plugin. It has a
// Config to store the configuration
// Plugin the plugin function
// suspended indicates that the plugin is temporarely suspended
type MagicZone struct {
	Config
	Plugin    soundtouch.PluginFunc
	suspended bool
}

// NewCollector creates a new Collector plugin with the configuration
func NewCollector(config Config) (d *MagicZone) {
	d = &MagicZone{}
	d.Config = config

	mLogger := log.WithFields(log.Fields{
		"Plugin": name,
	})

	mLogger.Debugf("Initialised\n")

	return d
}

// Name returns the plugin name
func (d *MagicZone) Name() string {
	return name
}

// Description returns a string explaining the purpose of this plugin
func (d *MagicZone) Description() string { return description }

// SampleConfig returns text explaining how plugin should be configured
func (d *MagicZone) SampleConfig() string { return sampleConfig }

// Terminate indicates that no further plugin will be executed on this speaker
func (d *MagicZone) Terminate() bool { return false }

// Disable temporarely the execution of the plugin
func (d *MagicZone) Disable() { d.suspended = true }

// Enable temporarely the execution of the plugin
func (d *MagicZone) Enable() { d.suspended = false }

// IsEnabled returns true if the plugin is not suspened
func (d *MagicZone) IsEnabled() bool { return !d.suspended }

// Execute runs the plugin with the given parameter
func (d *MagicZone) Execute(pluginName string, update soundtouch.Update, speaker soundtouch.Speaker) {
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
	if !(np.PlayStatus == soundtouch.PlayState) {
		mLogger.Debugln("PlayStatus != PlayState --> Done!")
		return
	}

	mLogger.Debugln("PlayStatus == PlayState --> Continuing")

	if !(np.StreamType == soundtouch.RadioStreaming) {
		mLogger.Debugln("StreamType != RadioStreaming. --> Done!")
		return
	}
	mLogger.Debugln("StreamType == RadioStreaming --> Continuing")
	compatibleStreamers := make([]soundtouch.Speaker, 0)
	for _, aKnownDevice := range soundtouch.GetKnownDevices() {
		mLogger.Tracef("aKnwonDevice: %s", aKnownDevice.Name())
		if speaker.DeviceInfo.DeviceID == aKnownDevice.DeviceInfo.DeviceID {
			continue
		}
		snp, _ := aKnownDevice.NowPlaying()
		if np.Content == snp.Content {
			mLogger.Debugln("Found other speaker streaming the same content --> Adding & Continuing")
			compatibleStreamers = append(compatibleStreamers, *aKnownDevice)
		}
	}

	if len(compatibleStreamers) == 0 {
		mLogger.Debugln("No other speaker found streaming the same content --> Done!")
		return // as there are no other speakers streaming the same content
	}

	// 1. Check: Already any zones defined?
	mLogger.Traceln("Are there already any zones in any compatibleStreamer?")
	for _, c := range compatibleStreamers {
		mLogger.Tracef("A compatibleStreamer: %s", c.Name())
		if c.HasZone() {
			mLogger.Traceln("Streamer is in a zone")
			// search for the one server that is indicated as master "zone.master == c.ownDeviceId"
			zone, _ := c.GetZone()
			if zone.Master == c.DeviceInfo.DeviceID {
				mLogger.Traceln("CompatbileStreamer is zoneMaster")
				if !speaker.IsSpeakerMember(zone.Members) {
					mLogger.Infof("Adding myself to master %v zone.\n", zone.Master)
					newZone := soundtouch.NewZone(c, speaker)
					c.AddZoneSlave(newZone)
					soundtouch.DumpZones(mLogger, c)
					mLogger.Debugln("Done!")
					return
				}
			}
		}
	}

	choosenAsNewMaster := compatibleStreamers[0]
	if !choosenAsNewMaster.HasZone() {
		newZone := soundtouch.NewZone(choosenAsNewMaster, speaker)
		mLogger.Infof("Creating new zone with %v as master.\n", newZone.Master)
		choosenAsNewMaster.SetZone(newZone)
		soundtouch.DumpZones(mLogger, choosenAsNewMaster)
		return
	}

}
