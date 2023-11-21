setup for sdcore
1. install go

build sdcore as plugin:
1. cd sdcore
2. go mod init sdcore
4. cd ./plugin
4. go build -buildmode=plugin -o sdcore.so work.go

usage for sdcore:
selfdata write/read --plugin=sdcore.so|sdcore.dll