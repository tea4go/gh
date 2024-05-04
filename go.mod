module github.com/tea4go/gh

go 1.18

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/mdlayher/arp v0.0.0-20220512170110-6706a2966875
	github.com/miekg/dns v1.1.56
	github.com/shiena/ansicolor v0.0.0-20230509054315-a9deabde6e02
	golang.org/x/net v0.15.0
	golang.org/x/text v0.13.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/josharian/native v1.0.0 // indirect
	github.com/mdlayher/ethernet v0.0.0-20220221185849-529eae5b6118 // indirect
	github.com/mdlayher/packet v1.0.0 // indirect
	github.com/mdlayher/socket v0.2.1 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
)
