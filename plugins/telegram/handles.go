package telegram

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/theovassiliou/soundtouch-golang"
	"golang.org/x/exp/slices"
	tb "gopkg.in/tucnak/telebot.v2"
)

// assertSender returns false in case user is not authorized
func (d *Bot) assertSender(sender *tb.User) bool {
	return slices.Contains(d.Config.AuthorizedSender, strconv.FormatInt(sender.ID, 10))
}

// /status [speakerName]
func (d *Bot) status(m *tb.Message) {
	if !d.assertSender(m.Sender) {
		d.bot.Send(m.Sender, fmt.Sprintf("%s (%v) not authorized. Use /authorize (authKey)", m.Sender.Username, m.Sender.ID))
		return
	}

	text := m.Text
	speakers := strings.Split(text, " ")
	speakers = speakers[1:]

	gkd := soundtouch.GetKnownDevices()
	var b strings.Builder
	if len(speakers) == 0 {
		b.WriteString("List of Soundtouch Devices\n")
		for sName := range gkd {
			speaker := soundtouch.GetSpeakerByDeviceId(sName)
			fmt.Fprintf(&b, "Device %s-%s with IP %s\n", speaker.Name(), speaker.DeviceID(), speaker.IP)

		}
	} else {
		for _, sName := range speakers {
			speaker := soundtouch.GetSpeakerByName(sName)
			if speaker != nil {
				fmt.Fprintf(&b, "Device %s-%s with IP %s\n", speaker.Name(), speaker.DeviceID(), speaker.IP)
				fmt.Fprintf(&b, " isPoweredOn(): %v\n", speaker.IsPoweredOn())
				fmt.Fprintf(&b, " isAlive(): %v\n", speaker.IsAlive())
				zone, _ := speaker.GetZone()

				fmt.Fprintf(&b, " isMaster(): %v\n", speaker.IsMaster())
				fmt.Fprintln(&b, "  zone.Master: ", zone.Master)
				fmt.Fprintln(&b, "  zone.SenderIPAddress: ", zone.SenderIPAddress)
				fmt.Fprintln(&b, "  zone.SenderIsMaster: ", zone.SenderIsMaster)
				fmt.Fprintln(&b, "  zone.Members: ", zone.Members)

				if speaker.IsAlive() {
					np, _ := speaker.NowPlaying()
					np.Raw = []byte{}
					fmt.Fprintf(&b, "Now Playing: %#v", np)
				}
			} else {
				fmt.Fprintf(&b, "Could not find speaker %v", sName)
			}
		}
	}
	d.bot.Send(m.Sender, b.String())
}

// /authorize [authkey]
func (d *Bot) authorize(m *tb.Message) {

	mLogger := log.WithFields(log.Fields{
		"Plugin": "TelegramBot",
		"Handle": "authorize/",
	})

	mLogger.Debugf("The bot: %v", d)
	authKey := d.Config.AuthKey

	if authKey == "" {
		d.bot.Send(m.Sender, "Authorization temporary disabled")
		return
	}

	text := m.Text
	authParam := strings.Split(text, " ")
	if len(authParam) >= 2 && authKey == authParam[1] {
		d.Config.AuthorizedSender = append(d.Config.AuthorizedSender, strconv.FormatInt(m.Sender.ID, 10))
		d.bot.Send(m.Sender, "Authorization granted")
		return
	} else if len(authParam) < 2 {
		d.bot.Send(m.Sender, "authorization key mising or wrong")
		return
	}

	d.bot.Send(m.Sender, fmt.Sprintf("Could not authorize with key %v", authParam[1]))

}
