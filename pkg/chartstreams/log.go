package chartstreams

import (
	"os"

	"github.com/sirupsen/logrus"
)

func SetLogLevel(level string) {
	logrus.SetOutput(os.Stdout)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(lvl)
}
