package episodecollector

import (
	"reflect"

	scribble "github.com/nanobox-io/golang-scribble"
	log "github.com/sirupsen/logrus"
	"github.com/theovassiliou/soundtouch-golang"
	"golang.org/x/exp/slices"
)

var name = "EpisodeCollector"

const description = "Collects episodes for specific artists"

const sampleConfig = `
## Enabling the episodeCollector plugin
# [episodeCollector]

## speakers for which episodes should be stores. If empty, all 
# speakers = ["Office", "Kitchen"]

## For which artists to collect the episodes
## all if empty
# artists = ["Drei Frageezeichen","John Sinclair"] 

## database contains the directory name for the episodes database
# database = "episode.db"

`

// Config contains the configuration of the plugin
// Speakers list of SpeakerNames the handler is added. All if empty
// Artists a list of artists for which episodes should be collected
type Config struct {
	Speakers []string `toml:"speakers"`
	Artists  []string `toml:"artists"`
	Database string   `toml:"database"`
}

// Collector describes the plugin. It has a
// Config to store the configuration
// Plugin the plugin function
// suspended indicates that the plugin is temporarely suspended
type Collector struct {
	Config
	Plugin     soundtouch.PluginFunc
	suspended  bool
	scribbleDb *scribble.Driver
}

// NewCollector creates a new Collector plugin with the configuration
func NewCollector(config Config) (d *Collector) {
	d = &Collector{}
	d.Config = config

	mLogger := log.WithFields(log.Fields{
		"Plugin": name,
	})

	mLogger.Debugf("Initialised\n")
	mLogger.Tracef("Scanning for: %v\n", d.Artists)

	db, err := scribble.New(d.Database, nil)
	if err != nil {
		log.Fatalf("Error with database. %s", err)
	}
	d.scribbleDb = db

	return d
}

// Name returns the plugin name
func (d *Collector) Name() string {
	return name
}

// Description returns a string explaining the purpose of this plugin
func (d *Collector) Description() string { return description }

// SampleConfig returns text explaining how plugin should be configured
func (d *Collector) SampleConfig() string { return sampleConfig }

// Terminate indicates that no further plugin will be executed on this speaker
func (d *Collector) Terminate() bool { return false }

// Disable temporarely the execution of the plugin
func (d *Collector) Disable() { d.suspended = true }

// Enable temporarely the execution of the plugin
func (d *Collector) Enable() { d.suspended = false }

// IsEnabled returns true if the plugin is not suspened
func (d *Collector) IsEnabled() bool { return !d.suspended }

// Execute runs the plugin with the given parameter
func (d *Collector) Execute(pluginName string, update soundtouch.Update, speaker soundtouch.Speaker) {
	if !(update.Is("NowPlaying") || update.Is("Volume")) {
		// UpdateMessageType not needed. Ignoring.
		return
	}

	if len(d.Speakers) > 0 && !slices.Contains(d.Speakers, speaker.Name()) {
		// Speaker not handled. Ignoring.
		return
	}

	mLogger := log.WithFields(log.Fields{
		"Plugin":        name,
		"Speaker":       speaker.Name(),
		"UpdateMsgType": reflect.TypeOf(update.Value).Name(),
	})
	mLogger.Debugln("Executing", pluginName)

	artist := update.Artist()
	album := update.Album()

	if !slices.Contains(d.Config.Artists, artist) || !update.HasContentItem() {
		mLogger.Debugf("Ignoring album: %s\n", album)
		return
	}

	mLogger.Infof("Found album: %v\n", album)
	readAlbumDB(d.scribbleDb, album, update)
}
