module github.com/lmatte7/meshtastic-go

go 1.16

replace github.com/lmatte7/go-meshtastic-protobufs => ./go-meshtastic-protobufs

require (
	github.com/jacobsa/go-serial v0.0.0-20180131005756-15cf729a72d4
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07 // indirect
	golang.org/x/sys v0.0.0-20210331175145-43e1dd70ce54 // indirect
	google.golang.org/protobuf v1.26.0
)
