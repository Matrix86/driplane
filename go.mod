module github.com/Matrix86/driplane

go 1.13

require (
	github.com/alecthomas/participle v0.2.1
	github.com/asaskevich/EventBus v0.0.0-20180315140547-d46933a94f05
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/dghubble/go-twitter v0.0.0-20190305084156-0022a70e9bee
	github.com/dghubble/oauth1 v0.5.0
	github.com/dghubble/sling v1.2.0
	github.com/evilsocket/islazy v1.10.4
	github.com/fsnotify/fsnotify v1.4.7
	github.com/google/go-querystring v1.0.0
	github.com/hpcloud/tail v1.0.0
	golang.org/x/sys v0.0.0-20190508100423-12bbe5a7a520
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
)

replace gopkg.in/fsnotify.v1 v1.4.7 => github.com/fsnotify/fsnotify v1.4.7
