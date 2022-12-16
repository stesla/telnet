package telnet

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
)

func TestLogf(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := NewMockLog(ctrl)
	log.EXPECT().Printf("[%s] format %s", WARN, "baz")
	log.EXPECT().Printf("[%s] format %s", ERROR, "quux")

	logger := NewLogLogger(log)
	logger.Logf(DEBUG, "format %s", "foo")
	logger.Logf(INFO, "format %s", "bar")
	logger.Logf(WARN, "format %s", "baz")
	logger.Logf(ERROR, "format %s", "quux")
}

func TestSetLevel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := NewMockLog(ctrl)
	log.EXPECT().Printf("[%s] foo", DEBUG)

	logger := NewLogLogger(log)
	logger.SetLevel(DEBUG)
	logger.Logf(DEBUG, "foo")
}
