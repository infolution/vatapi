#!/bin/sh
GO=/usr/local/go/bin/go
$GO build src/vatapi-server.go
sudo cp vatapi-server /opt/vatapi/
sudo cp taxes-cleaned.csv /opt/vatapi/