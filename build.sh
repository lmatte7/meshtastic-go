env GOOS=darwin GOARCH=amd64 go build -o builds/meshtastic_go_osx ./cmd/meshtastic-go
env GOOS=linux GOARCH=arm GOARM=5 go build -o builds/meshtastic_go_linux_arm ./cmd/meshtastic-go
env GOOS=linux GOARCH=arm64 go build -o builds/meshtastic_go_linux_arm64 ./cmd/meshtastic-go
env GOOS=linux GOARCH=amd64 go build -o builds/meshtastic_go_linux_amd64 ./cmd/meshtastic-go
env GOOS=linux GOARCH=386 go build -o builds/meshtastic_go_linux_amd64 ./cmd/meshtastic-go
env GOOS=windows GOARCH=amd64 go build -o builds ./cmd/meshtastic-go
