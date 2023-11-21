# !/bin/bash
nowtime=$(date "+%Y%m%d%H%M%S")
workDir=/working/release/rsapp/selfdata

if [[ "$1" =~ "selfdata_" ]] ;then
  mkdir -p $workDir/build/$nowtime/datas
  cd $workDir/build/$nowtime
  cp $workDir/publish/$1 $workDir/build/$nowtime
  cp $workDir/build/build.go $workDir/build/$nowtime
  $workDir/build/rspub --app=selfdata --dpath=./datas 1>/dev/null 2>&1
  cp -r $workDir/publish/web $workDir/build/$nowtime/datas
  $workDir/build/rice append --exec $1 1>/dev/null 2>&1
  chmod +x $workDir/build/$nowtime/$1
  echo "path:$workDir/build/$nowtime/$1"
fi
