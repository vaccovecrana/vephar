#!/bin/bash
cd ui/
npm install
npm run build
cp ./res/favicon.ico ../srv/dist/

cd ..
cd srv/
go build
