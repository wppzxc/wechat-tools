## build
``` powershell
rsrc.exe -manifest wechat-tools.exe.manifest -arch amd64 -ico ./assets/img/icon.ico -o rsrc.syso
go build -ldflags="-H windowsgui -X github.com/wppzxc/wechat-tools/version.version='$version'"
``` 