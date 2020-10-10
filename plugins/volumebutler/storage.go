package volumebutler

import (
	"reflect"
	"time"

	scribble "github.com/nanobox-io/golang-scribble"
	log "github.com/sirupsen/logrus"
	"github.com/theovassiliou/soundtouch-golang"
)

type DbEntry struct {
	ContentItem soundtouch.ContentItem
	AlbumName   string
	Volume      int
	DeviceID    string
	LastUpdated time.Time
}

func readDB(sdb *scribble.Driver, album string, currentAlbum *DbEntry) *DbEntry {
	if currentAlbum == nil {
		currentAlbum = &DbEntry{}
	}
	sdb.Read("All", album, &currentAlbum)
	return currentAlbum
}

func readAlbumDB(sdb *scribble.Driver, album string, updateMsg soundtouch.Update) *DbEntry {

	storedAlbum := readDB(sdb, album, &DbEntry{})

	if storedAlbum.AlbumName == "" {
		// no, write this into the database
		storedAlbum.AlbumName = album
		// HYPO: We are in observation window, then the current volume could also
		// be a good measurement
		storedAlbum.DeviceID = updateMsg.DeviceID
		storedAlbum.ContentItem = updateMsg.ContentItem()
		writeDB(sdb, "All", album, storedAlbum)
	}
	return storedAlbum
}

func writeDB(sdb *scribble.Driver, collection, album string, storedAlbum *DbEntry) {
	storedAlbum.LastUpdated = time.Now()
	sdb.Write(collection, album, &storedAlbum)
}

func WriteDB(sdb *scribble.Driver, speakerName, album string, storedAlbum *DbEntry) {
	storedAlbum.LastUpdated = time.Now()
	sdb.Write(speakerName, album, &storedAlbum)
	sdb.Write("All", album, &storedAlbum)
}

func ReadAlbumDB(sdb *scribble.Driver, album string, updateMsg soundtouch.Update) *DbEntry {

	speaker := soundtouch.GetSpeaker(updateMsg)
	if speaker == nil {
		return nil
	}

	mLogger := log.WithFields(log.Fields{
		"Plugin":        name,
		"Speaker":       speaker.Name(),
		"UpdateMsgType": reflect.TypeOf(updateMsg.Value).Name(),
	})

	storedAlbum := ReadDB(sdb, speaker.Name(), album, &DbEntry{})

	if storedAlbum.AlbumName == "" {
		mLogger.Infof("Album %s not yet known. Reading volume.", storedAlbum.AlbumName)
		// no, write this into the database
		retrievedVol, _ := speaker.Volume()
		mLogger.Infof("Volume is %d", retrievedVol.ActualVolume)
		storedAlbum.AlbumName = album
		// HYPO: We are in observation window, then the current volume could also
		// be a good measurement
		storedAlbum.Volume = retrievedVol.TargetVolume
		storedAlbum.DeviceID = updateMsg.DeviceID
		storedAlbum.LastUpdated = time.Now()
		storedAlbum.ContentItem = updateMsg.ContentItem()
		writeDB(sdb, speaker.Name(), album, storedAlbum)
		writeDB(sdb, "ALL", album, storedAlbum)
	}
	return storedAlbum
}

// ReadDB returns a databaseEntry for a given Album, or an empty databaseEntry if collection has no album stored
func ReadDB(sdb *scribble.Driver, collection string, album string, currentAlbum *DbEntry) *DbEntry {
	if currentAlbum == nil {
		currentAlbum = &DbEntry{}
	}
	sdb.Read(collection, album, &currentAlbum)
	return currentAlbum
}

func (db *DbEntry) calcNewVolume(currVolume int) int {
	oldVol := db.Volume
	if oldVol == 0 {
		oldVol = currVolume
	}
	return (oldVol + currVolume) / 2
}
