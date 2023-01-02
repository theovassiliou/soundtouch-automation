module github.com/theovassiliou/soundtouch-master

go 1.18

replace github.com/theovassiliou/soundtouch-golang => ../soundtouch-golang

require (
	github.com/gorilla/websocket v1.5.0
	github.com/influxdata/toml v0.0.0-20180607005434-2a2e3012f7cf
	github.com/jpillora/opts v1.2.0
	github.com/nanobox-io/golang-scribble v0.0.0-20190309225732-aa3e7c118975
	github.com/sirupsen/logrus v1.7.0
	github.com/theovassiliou/soundtouch-golang v0.0.0-20200612082517-79e25fc76cbd
	golang.org/x/exp v0.0.0-20221230185412-738e83a70c30
	gopkg.in/tucnak/telebot.v2 v2.3.4
)

require (
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/mdns v1.0.3 // indirect
	github.com/jcelliott/lumber v0.0.0-20160324203708-dd349441af25 // indirect
	github.com/miekg/dns v1.1.27 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/posener/complete v1.2.2-0.20190308074557-af07aa5181b3 // indirect
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/sys v0.2.0 // indirect
)
