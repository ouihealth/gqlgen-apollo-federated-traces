#!/bin/sh
curl -s https://raw.githubusercontent.com/apollographql/apollo-server/main/packages/apollo-reporting-protobuf/src/reports.proto | sed -E 's/ +\[.*$/;/g' > reports.proto
protoc --go_opt=Mreports.proto=github.com/ouihealth/gqlgen-apollo-reporting/reports --go_out=. reports.proto
