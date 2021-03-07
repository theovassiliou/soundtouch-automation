# InfluxConnector

The influxconnector plugin enables the recording of various soundtouch message updates
to an existing [InfluxDB](https://www.influxdata.com) installation.

The plugin is enabled by including an `[influxDB]` section in your
configuration-toml file.

```toml
[influxDB]
## speakers for which messages should be logged. If empty, all 
speakers = ["Office", "Kitchen"]

## log_messages describes the message types to be logged
## one or more of "ConnectionStateUpdated", "NowPlaying", "Volume"
## all if empty
log_messages = ["ConnectionStateUpdated", "NowPlaying", "Volume"] 

## URL of the InfluxDB
influxURL = "http://influxdb:8086"

## Database where to store the events
database = "soundtouch"

#
## dry_run indicates that the plugin dumps lineporoto for the influxDB conncetion
## as curl statement. 
# dry_run = true
## 
```

If a connection to the provided influxDB can not be established, the plugin will be disabled after some 20 retries.
