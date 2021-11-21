env GOOS=darwin GOARCH=amd64 go build -o builds/meshtastic_go_osx github.com/lmatte7/meshtastic-go
env GOOS=linux GOARCH=arm GOARM=5 go build -o builds/meshtastic_go_linux_arm github.com/lmatte7/meshtastic-go
env GOOS=linux GOARCH=arm64 go build -o builds/meshtastic_go_linux_arm64 github.com/lmatte7/meshtastic-go
env GOOS=linux GOARCH=amd64 go build -o builds/meshtastic_go_linux_amd64 github.com/lmatte7/meshtastic-go
env GOOS=linux GOARCH=386 go build -o builds/meshtastic_go_linux_amd64 github.com/lmatte7/meshtastic-go
env GOOS=windows go build -o builds github.com/lmatte7/meshtastic-go