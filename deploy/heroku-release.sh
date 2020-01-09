cd /tmp
wget https://storage.googleapis.com/golang/go1.11.8.linux-amd64.tar.gz
/bin/tar xvf go1.11.8.linux-amd64.tar.gz
/tmp/go/bin/go version
cd -
/tmp/go/bin/go ./cmd/migrate/migrate.go