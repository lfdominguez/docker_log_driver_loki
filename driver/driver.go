package driver

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"path"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/loggerutils"
	protoio "github.com/gogo/protobuf/io"
	"github.com/pkg/errors"
	"github.com/tonistiigi/fifo"
)

const (
	//log-opt
	lokihost = "loki-host"
	lokiport = "loki-port"
)

type Driver struct {
	mu   sync.Mutex
	logs map[string]*logPair
}

type logPair struct {
	active  bool
	file    string
	info    logger.Info
	logLine jsonLogLine
	stream  io.ReadCloser
}

func validateDriverOpt(loggerInfo logger.Info) error {
	config := loggerInfo.Config

	for opt := range config {
		switch opt {
		case lokihost, lokiport:
		case "labels":
		case "env":
		case "env-regex":
		case "tag":
		default:
			return fmt.Errorf("wrong log-opt: '%s' - %s\n", opt, loggerInfo.ContainerID)
		}
	}
	_, ok := config[lokihost]
	if !ok {
		return fmt.Errorf("Loki host is required. config: %v+\n", config)
	}

	_, ok = config[lokiport]
	if !ok {
		return fmt.Errorf("Loki host port is required\n")
	}

	return nil
}

func NewDriver() *Driver {
	return &Driver{
		logs: make(map[string]*logPair),
	}
}

func (d *Driver) StartLogging(file string, logCtx logger.Info) error {
	d.mu.Lock()
	if _, exists := d.logs[path.Base(file)]; exists {
		d.mu.Unlock()
		return fmt.Errorf("logger for %q already exists", file)
	}
	d.mu.Unlock()

	logrus.WithField("id", logCtx.ContainerID).WithField("file", file).Info("Start logging")
	stream, err := fifo.OpenFifo(context.Background(), file, syscall.O_RDONLY, 0700)
	if err != nil {
		return errors.Wrapf(err, "error opening logger fifo: %q", file)
	}

	err = validateDriverOpt(logCtx)
	if err != nil {
		return errors.Wrap(err, "error in one of the logger options")
	}

	tag, err := loggerutils.ParseLogTag(logCtx, loggerutils.DefaultTemplate)
	if err != nil {
		return errors.Wrapf(err, "error reading log tags\n")
	}

	extra, err := logCtx.ExtraAttributes(nil)
	if err != nil {
		return errors.Wrapf(err, "error reading extra attributes\n")
	}

	hostname, err := logCtx.Hostname()
	if err != nil {
		return errors.Wrapf(err, "error reading hostname\n")
	}

	logLine := jsonLogLine{
		ContainerId:   logCtx.FullID(),
		ContainerName: logCtx.Name(),
		StackName:     extra["com.docker.stack.namespace"],
		ServiceName:   extra["com.docker.swarm.service.name"],
		ImageId:       logCtx.ImageFullID(),
		ImageName:     logCtx.ImageName(),
		Command:       logCtx.Command(),
		Tag:           tag,
		Extra:         extra,
		Host:          hostname,
	}

	lp := &logPair{true, file, logCtx, logLine, stream}

	d.mu.Lock()
	d.logs[path.Base(file)] = lp
	d.mu.Unlock()

	go consumeLog(lp)

	return nil
}

func (d *Driver) StopLogging(file string) error {
	logrus.WithField("file", file).Info("Stop logging")
	d.mu.Lock()
	lp, ok := d.logs[path.Base(file)]
	if ok {
		lp.active = false
		delete(d.logs, path.Base(file))
	} else {
		logrus.WithField("file", file).Errorf("Failed to stop logging. File %q is not active", file)
	}
	d.mu.Unlock()
	return nil
}

func shutdownLogPair(lp *logPair) {
	if lp.stream != nil {
		lp.stream.Close()
	}

	lp.active = false
}

func consumeLog(lp *logPair) {
	var buf logdriver.LogEntry

	dec := protoio.NewUint32DelimitedReader(lp.stream, binary.BigEndian, 1e6)
	defer dec.Close()
	defer shutdownLogPair(lp)

	logrus.WithField("id", lp.info.ContainerID).Debug("sending log to Loki server")

	for {
		err := dec.ReadMsg(&buf)
		if err != nil {
			if err == io.EOF {
				logrus.WithField("id", lp.info.ContainerID).WithError(err).Debug("shutting down logger goroutine due to file EOF")
				return
			} else {
				logrus.WithField("id", lp.info.ContainerID).WithError(err).Warn("error reading from FIFO, trying to continue")
				dec = protoio.NewUint32DelimitedReader(lp.stream, binary.BigEndian, 1e6)
				continue
			}
		}

		err = logMessageToLoki(lp, buf.Line)
		if err != nil {
			logrus.WithField("id", lp.info.ContainerID).WithError(err).Warn("error logging message, dropping it and continuing")
		}

		buf.Reset()
	}
}

func (d *Driver) ReadLogs(info logger.Info, config logger.ReadConfig) (io.ReadCloser, error) {
	return nil, nil
}
