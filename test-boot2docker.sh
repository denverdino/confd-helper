export GOARCH="amd64"
export GOOS="linux"
echo "Building ..."
go build -o confd-helper-linux
boot2docker ssh "chmod +x $(pwd)/confd-helper-linux"
echo "Start testing"
boot2docker ssh "$(pwd)/confd-helper-linux exec --container=5b08d3523b24 -c ls"

