export GOARCH="amd64"
export GOOS="linux"
echo "Building ..."
go build -o confd-helper-linux
boot2docker ssh "chmod +x $(pwd)/confd-helper-linux"
echo "Start testing"
CID=$(docker run -d jadetest.cn.ibm.com:5000/nginx)
boot2docker ssh "$(pwd)/confd-helper-linux exec --container=$CID -c 'touch /abc.txt'"
docker kill $CID
docker rm $CID
