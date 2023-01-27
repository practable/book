#!/bin/bash
mkdir ./pprof || true #ignore error
echo "capturing 30 sec CPU profile"
echo "useful pprof commands: top10, web"
curl http://localhost:6060/debug/pprof/profile > ./pprof/profile
go tool pprof ./pprof/profile
