package log

import (
	"github.com/labstack/echo/v4/middleware"
	"io"
	"os"
)

var DefaultLogConfig = middleware.LoggerConfig{
	Skipper: middleware.DefaultSkipper,
	Format: `"time":"${time_rfc3339_nano}" "method":"${method}" "uri":"${uri}" "status":${status} "error":"${error}" ` +
		`"latency":${latency} "latency_human":"${latency_human}" "bytes_in":${bytes_in} "bytes_out":${bytes_out}` + "\n",
	CustomTimeFormat: "2006-01-02 15:04:05.00000",
	Output:           GetLogFileWriter("./wechat-tools.log"),
}

func GetLogFileWriter(path string) io.Writer {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return os.Stdout
	}
	return file
}
