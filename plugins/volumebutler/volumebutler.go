package volumebutler

import (
	"reflect"
	"time"

	scribble "github.com/nanobox-io/golang-scribble"
	log "github.com/sirupsen/logrus"
	soundtouch "github.com/theovassiliou/soundtouch-golang"
	"golang.org/x/exp/slices"
)

var name = "volumeButler"

const sampleConfig = `
## Enabling the volumeButler plugin
# [volumeButler]

## speakers for which volumeButler will handle volumes. None if empty. 
# speakers = ["Office", "Kitchen"]

## For which artists volumes should be handled
## all if empty
# artists = ["Drei Frageezeichen","John Sinclair"] 

## database contains the directory name for the episodes database
# database = "episode.db"
`

const description = "Automatically adjust sets volume based on listening history."

// Config contains the configuration of the plugin
// Speakers list of SpeakerNames the handler is added. All if empty
// Artists a list of artists for which episodes should be collected
type Config struct {
	Speakers []string `toml:"speakers"`
	Artists  []string `toml:"artists"`
	Database string   `toml:"database"`
}

// VolumeButler describes the plugin. It has a
// Config to store the configuration
// Plugin the plugin function
// suspended indicates that the plugin is temporarely suspended
// scribbleDB a link to the volumes database
type VolumeButler struct {
	Config
	Plugin     soundtouch.PluginFunc
	suspended  bool
	scribbleDb *scribble.Driver
}

// NewVolumeButler creates a new Collector plugin with the configuration
func NewVolumeButler(config Config) (d *VolumeButler) {
	d = &VolumeButler{}
	if config.Database == "" {
		return d
	}
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

	mLogger.Debugf("Initialised database: %v\n", db)
	d.scribbleDb = db

	return d
}

// Name returns the plugin name
func (vb *VolumeButler) Name() string {
	return name
}

// Description returns a string explaining the purpose of this plugin
func (vb *VolumeButler) Description() string { return description }

// SampleConfig returns text explaining how plugin should be configured
func (vb *VolumeButler) SampleConfig() string { return sampleConfig }

// Terminate indicates that no further plugin will be executed on this speaker
func (vb *VolumeButler) Terminate() bool { return false }

// Disable temporarely the execution of the plugin
func (vb *VolumeButler) Disable() { vb.suspended = true }

// Enable temporarely the execution of the plugin
func (vb *VolumeButler) Enable() { vb.suspended = false }

// IsEnabled returns true if the plugin is not suspened
func (vb *VolumeButler) IsEnabled() bool { return !vb.suspended }

// Execute runs the plugin with the given parameter
func (vb *VolumeButler) Execute(pluginName string, update soundtouch.Update, speaker soundtouch.Speaker) {

	typeName := reflect.TypeOf(update.Value).Name()
	mLogger := log.WithFields(log.Fields{
		"Plugin":        name,
		"Speaker":       speaker.Name(),
		"UpdateMsgType": reflect.TypeOf(update.Value).Name(),
	})
	mLogger.Debugln("Executing", pluginName)

	if len(vb.Speakers) == 0 || !slices.Contains(vb.Speakers, speaker.Name()) {
		mLogger.Debugln("Speaker not handled. --> Done!")
		return
	}

	if !(update.Is("NowPlaying") || update.Is("Volume")) {
		mLogger.Debugf("Ignoring %s. --> Done!\n", typeName)
		return
	}

	artist := update.Artist()
	album := update.Album()

	if !slices.Contains(vb.Config.Artists, artist) || !update.HasContentItem() {
		mLogger.Debugf("Ignoring album %s from %s\n", album, artist)
		return
	}
	mLogger.Infof("Found album %s from %s\n", album, artist)
	// time window independend
	// Do we know this album already?  - read from database
	storedAlbum := ReadAlbumDB(vb.scribbleDb, album, update)

	// time window and speaker depended
	// 		if available for this album
	//			set the volume
	if storedAlbum.Volume != 0 && time.Now().After(storedAlbum.LastUpdated.Add(20*time.Minute)) {
		// Only setting a volume if it was last update more than 20 minutes ago
		mLogger.Infof("Stored volume was set more than 20minutes ago\n")
		mLogger.Infof("Setting volume to %d\n", storedAlbum.Volume)
		speaker.SetVolume(storedAlbum.Volume)
	}

	// wait for a minute and process last volume observed
	// construct the mean value of current and past volumes
	// store the update value
	mLogger.Infof("Going to sleep for 60s\n")
	time.Sleep(60 * time.Second)

	// clear message and keep last volume update
	mLogger.Infof("Scanning for Volume\n")
	lastVolume := ScanForVolume(&speaker)
	ReadDB(vb.scribbleDb, speaker.Name(), album, storedAlbum)
	if lastVolume != nil {
		storedAlbum.Volume = storedAlbum.calcNewVolume(lastVolume.TargetVolume)
		mLogger.Infof("writing volume to %v\n", storedAlbum.Volume)
		vb.scribbleDb.Write(speaker.Name(), album, &storedAlbum)
	}
}

// ScanForVolume listens for Volume updates and returns the latest received Volume, when the
// channel is closed
func ScanForVolume(spk *soundtouch.Speaker) *soundtouch.Volume {
	var lastVolume *soundtouch.Volume
	var mLogger *log.Entry
	for scanMsg := range spk.WebSocketCh {
		typeName := reflect.TypeOf(scanMsg.Value).Name()
		mLogger = log.WithFields(log.Fields{
			"Plugin":        name,
			"Speaker":       spk.Name(),
			"UpdateMsgType": typeName,
		})
		if scanMsg.Is("Volume") {
			aVol, _ := scanMsg.Value.(soundtouch.Volume)
			lastVolume = &aVol
			mLogger.Printf("Ignoring! Volume: %#v", lastVolume)
		} else {
			mLogger.Debugf("Ignoring!! %s\n", typeName)
		}
		if len(spk.WebSocketCh) == 0 {
			break
		}
	}

	if lastVolume != nil {
		mLogger.Infof("lastVolume was %d\n", lastVolume.ActualVolume)
	}
	return lastVolume
}
