# !/bin/bash

version="beta-1.0.1"
workDir=/home/jery/Desktop/dev/rshub

# init
if [ -n "$2" ] ;then
  version="$2"
fi
cd $workDir/product/rsapp/selfdata

# ./build.sh publish 1.0.0
if [[ "$1" = "publish" ]] ;then
    # build rspub
    cd $workDir/tools/rspub
    rm ./rspub*
    go build -ldflags="-w -s" -i -o $workDir/product/rsapp/selfdata/release/build/rspub
    upx --best --lzma $workDir/product/rsapp/selfdata/release/build/rspub
    cd $workDir/product/rsapp/selfdata

    # build go.rice
    cd $workDir/../tools/go.rice/rice
    rm ./rice*
    gox -osarch=linux/amd64 -ldflags="-w -s" -output="rice"
    mv ./rice* $workDir/product/rsapp/selfdata/release/build
    cd $workDir/product/rsapp/selfdata

  # build selfdata
  rm ./release/publish/selfdata*
  buildList=(darwin linux windows)
  for os in ${buildList[*]}
  do
    echo "start build for ${os}"
    cd $workDir/product/rsapp/selfdata
    rm ./selfdata*
    gox -os=${os} -arch=amd64 -ldflags="-w -s -X main.BuildID=$version" -output="{{.Dir}}_{{.OS}}_{{.Arch}}-$version"
    gox -os=${os} -arch=arm -ldflags="-w -s -X main.BuildID=$version" -output="{{.Dir}}_{{.OS}}_{{.Arch}}-$version"
    # 调整在官网通过rspub动态生成meta.sd, 动态生成下载链接
    # rice append --exec ./selfdata*
    mv ./selfdata* ./release/publish
  done
  cd $workDir/product/rsapp/selfdata/release

  # upx selfdata
  apps=`find ./publish -maxdepth 1 | grep "selfdata*"`
  for app in $apps
  do
    upx --best --lzma ${app}
  done

  # publish to rs
#  scp -r -P 5177 ./rsrelay refitor@refitself.cn:/home/refitor/rsapp/selfdata
#  scp -r -P 5177 ./build refitor@refitself.cn:/home/refitor/rsapp/selfdata
#  scp -r -P 5177 ./origin refitor@refitself.cn:/home/refitor/rsapp/selfdata
#  scp -r -P 5177 ./publish refitor@refitself.cn:/home/refitor/rsapp/selfdata
   scp -r -P 5177 ./build ./publish refitor@refitself.cn:/home/refitor/rsapp/selfdata
elif [[ "$1" = "web" ]] ;then
  cd web
  cnpm run build
else
  #rspub --app=selfdata --dpath=./datas
  go build -ldflags="-w -s -X main.DebugCode=xxxxxx" -i
  rice append --exec selfdata
  #upx --brute --best --lzma selfdata
  mv ./selfdata ./release
  cd ./release
fi