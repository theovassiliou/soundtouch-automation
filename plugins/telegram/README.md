# Telegram Connector

The telegram connector plugin enables management of the automator via the Telegram messaging systems

The plugin is enabled by including an `[telegram]` section in your
configuration-toml file.

```toml
[telegram]

## speakers for which messages should be logged. If empty, all 
speakers = ["Office", "Kitchen"]

## ignore_messages describes the message types to be ignored
## one or more of "ConnectionStateUpdated", "NowPlaying", "Volume"
## all if empty
ignore_messages = ["ConnectionStateUpdated"] 

## Telegram API Key
apiKey ="1292466187:AAG3O6QyfNSpEgNq5JrlpINz4w5z6bQIrk8"
authorizedSenders = ["999999", "888888"]
authKey = "ThisIsAVerySecretKey34abf77&"
```

In order to use the telegram plugin usefully you have register your "bot" with Telegram. 
For this you will need an `apiKey` which you have to provide in the respective field in the config-toml.

It is up to you to name and make the bot accessible to you and your authorized contacts. 

While everybody can see the bot in the Telegram world, only `authorizedSenders" will received insights in your Soundtouch systems.

You become an authorized used by either adding your ID in the field `authorizedSenders`. In addition a Telegram member can add himself to the list of authorized users by sending the message (command)

`/authorize $authKey$` to the bot. 

`$authKey$`is the key specified in the `authKey`field of the config-toml.

Please note that the authorization is not persistant and you have to reauthorize if you restart the automator. 

The following commands have been implemented so far

```text
/hello - You will receive your name and your userId back
/authorize [authKey] - You authorize yourself to the system
/stats [speakerName] - Get the status of your soundtouch system or from a specific speaker
```
