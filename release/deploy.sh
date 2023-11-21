nowtime=$(date "+%Y%m%d%H%M%S")
rm -rf ./backup
mkdir ./backup
for file in $(ls /home/refitor/rsapp/selfdata)
do
    echo $file
    if test -f $file
    then
        if [ -f $file ];then
            mv ./$file ./backup/$file
            mv /home/refitor/rsapp/selfdata/$file $pwd/
        fi
    fi
    if test -d $file
    then
        if [ -d $file ];then
            mv ./$file ./backup/$file
            mv /home/refitor/rsapp/selfdata/$file ./
        fi
    fi
done

chmod +x ./build/*
chmod +x ./publish/*
mkdir -p /working/release/rshost/web/download/selfdata/history/$nowtime
cp ./publish/* /working/release/rshost/web/download/selfdata/history/$nowtime
cp ./publish/selfdata_linux_amd64* /working/release/rshost/web/download/selfdata/selfdata_linux_amd64
cp ./publish/selfdata_linux_arm* /working/release/rshost/web/download/selfdata/selfdata_linux_arm
cp ./publish/selfdata_darwin_amd64* /working/release/rshost/web/download/selfdata/selfdata_darwin_amd64
cp ./publish/selfdata_windows_amd64* /working/release/rshost/web/download/selfdata/selfdata_windows_amd64.exe
