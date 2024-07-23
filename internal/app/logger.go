package app

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

func formatFilePath(path string) string {
	arr := strings.Split(path, "/")
	return arr[len(arr)-1]
}

func InitLogger() error {
	logrus.SetReportCaller(true)

	logrusLvl, err := logrus.ParseLevel(CFG.Logger.LogLvl)
	if err != nil {
		return err
	}

	txtFormatter := &logrus.TextFormatter{
		TimestampFormat:        "02-01-2006 15:04:05",
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		ForceColors:            true,
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			return "", fmt.Sprintf("%s:%d", formatFilePath(f.File), f.Line)
		},
	}

	logrus.SetLevel(logrusLvl)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(txtFormatter)

	return nil
}
