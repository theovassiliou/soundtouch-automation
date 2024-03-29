package telegram

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/theovassiliou/soundtouch-golang"
	"golang.org/x/exp/slices"
	tb "gopkg.in/tucnak/telebot.v2"
)

var name = "Telegram"

const sampleConfig = `
## Enabling logger plugin
# [telegram]

## speakers for which messages should be logged. If empty, all 
# speakers = ["Office", "Kitchen"]

## ignore_messages describes the message types to be ignored
## one or more of "ConnectionStateUpdated", "NowPlaying", "Volume"
## all if empty
# ignore_messages = ["ConnectionStateUpdated"] 

## Telegram API Key
# apiKey ="x:y"
# authorizedSenders = ["999999", "888888"]
# authKey = "secrectKey" 
`

const description = "Logs all update messages to telegram"

// Config contains the configuration of the plugin
// Speakers list of SpeakerNames the handler is added. All if empty
// IgnoreMessages a list of message types to be ignored
// APIKey for the telegram bot
type Config struct {
	Speakers         []string `toml:"speakers"`
	IgnoreMessages   []string `toml:"ignore_messages"`
	APIKey           string   `toml:"apiKey"`
	AuthorizedSender []string `toml:"authorizedSenders"`
	AuthKey          string   `toml:"authKey"`
}

// Bot describes the plugin. It has a
// Config to store the configuration
// Plugin the plugin function
type Bot struct {
	Config
	Plugin    soundtouch.PluginFunc
	suspended bool
	bot       *tb.Bot
}

// NewTelegramLogger creates a new Logger plugin with the configuration
func NewTelegramLogger(config Config) (d *Bot) {
	mLogger := log.WithFields(log.Fields{
		"Plugin": name,
	})

	d = &Bot{}
	if config.APIKey == "" {
		mLogger.Debug("No APIKey provided. Suspending plugin.")
		d.suspended = true
		return d
	}

	d.Config = config

	b, err := tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		URL: "https://api.telegram.org",

		Token:  config.APIKey,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}
	d.bot = b

	var (
		// Universal markup builders.
		menu     = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
		selector = &tb.ReplyMarkup{}

		// Reply buttons.
		btnHelp     = menu.Text("ℹ Help")
		btnSettings = menu.Text("⚙ Settings")

		// Inline buttons.
		//
		// Pressing it will cause the client to
		// send the bot a callback.
		//
		// Make sure Unique stays unique as per button kind,
		// as it has to be for callback routing to work.
		//
		btnPrev = selector.Data("⬅", "prev", "TEXT btnPrev")
		btnNext = selector.Data("➡", "next", "TEXT btnNext")
	)

	menu.Reply(
		menu.Row(btnHelp),
		menu.Row(btnSettings),
	)
	selector.Inline(
		selector.Row(btnPrev, btnNext),
	)

	// On reply button pressed (message)
	b.Handle(&btnHelp, func(m *tb.Message) {
		b.Send(m.Sender, fmt.Sprintf("Help %v!", m.Sender.FirstName))

	})

	// On inline button pressed (callback)
	b.Handle(&btnPrev, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{
			CallbackID: "",
			Text:       "testadfad ",
			ShowAlert:  false,
			URL:        "",
		},
		)
	})

	b.Handle("/status", func(m *tb.Message) {
		d.status(m)
	})

	b.Handle("/authorize", func(m *tb.Message) {
		d.authorize(m)
	})

	b.Handle("/hello", func(m *tb.Message) {
		b.Send(m.Sender, fmt.Sprintf("Hello %v(%v)!", m.Sender.FirstName, m.Sender.ID), menu)
	})
	s, _ := strconv.ParseInt(d.Config.AuthorizedSender[0], 10, 64)
	user := &tb.User{ID: s}
	b.Handle(tb.OnText, func(m *tb.Message) {
		mLogger.Infof("Recevived telegram message: %#v\n", m.Text)
		mLogger.Infof("  by: %v\n", m.Sender)
		if m.Sender.ID != s {
			b.Send(user, fmt.Sprintf("Recevived telegram message: %#v\n  by: %v\n", m.Text, m.Sender))
		}
		b.Send(user, fmt.Sprintf("Recevived telegram message: %#v\n  by: %v\n", m.Text, m.Sender))
	})

	mLogger.Debugf("Initialised\n")

	go b.Start()

	return d
}

// Name returns the plugin name
func (d *Bot) Name() string {
	return name
}

// Description returns a string explaining the purpose of this plugin
func (d *Bot) Description() string { return description }

// SampleConfig returns text explaining how plugin should be configured
func (d *Bot) SampleConfig() string { return sampleConfig }

// Terminate indicates that no further plugin will be executed on this speaker
func (d *Bot) Terminate() bool { return false }

// Disable temporarely the execution of the plugin
func (d *Bot) Disable() { d.suspended = true }

// Enable temporarely the execution of the plugin
func (d *Bot) Enable() { d.suspended = false }

// IsEnabled returns true if the plugin is not suspened
func (d *Bot) IsEnabled() bool { return !d.suspended }

// Execute runs the plugin with the given parameter
func (d *Bot) Execute(pluginName string, update soundtouch.Update, speaker soundtouch.Speaker) {
	if !d.IsEnabled() {
		return
	}
	if len(d.IgnoreMessages) > 0 && slices.Contains(d.IgnoreMessages, reflect.TypeOf(update.Value).Name()) {
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
	mLogger.Debugf("Executing %v on %v", pluginName, reflect.TypeOf(update.Value).Name())
}
