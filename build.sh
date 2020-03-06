#/bin/bash

#CGO_ENABLED=0 GOOS=linux GOARCH=amd64
#go build -a -o bin/userd ./src/main.go 

appname="userd"
nowtime=$(date +%Y%m%d%H%M)
dir=$(pwd)

###  initalize #####
echo "initalize ..."
echo "rm $dir/output"
rm -r $dir/output/ >/dev/null 2>&1
mkdir -p  output/bin
mkdir -p  output/conf
mkdir -p  output/log
mkdir -p  output/script
mkdir -p  output/fonts

###  build      ####
echo "start build ..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build  -o ./bin/$appname -tags springboot ./src/main.go || exit 1

### copy files  ####
echo "copy to destination dir"
mv ./bin/$appname               ./output/bin/$appname
cp ./conf/*.conf.test           ./output/conf/
cp ./load.sh                    ./output/
cp -r ./fonts/*                 ./output/fonts/

### tar ############
echo "tar ..."
cd output
tar -czf $appname.tar.gz ./bin ./conf ./fonts  ./log ./load.sh
mv ./$appname.tar.gz $dir/
rm -r $dir/output/

### done ############
echo "done,done,done"
