package logger

import (
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerConfig struct {
	Key        string
	Filename   string
	MaxSize    int
	MaxBackups int
	Level      logrus.Level
	MaxAge     int
}

var logger *logrus.Logger

func NewLogger(config LoggerConfig) *logrus.Logger {
	logger = logrus.New()
	logger.SetOutput(&lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   true,
	})
	logger.SetLevel(config.Level)
	return logger
}

func GetLogger() *logrus.Logger {
	if logger != nil {
		return logger
	}
	logger = logrus.New()
	return logger
}

const (
	reqidKey = "X-Reqid"
)

func NewGinMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		reqId := c.Request.Header.Get(reqidKey)
		if reqId == "" {
			reqId = GenReqId()
			c.Request.Header.Set(reqidKey, reqId)
		}
		h := c.Writer.Header()
		h.Set(reqidKey, reqId)

		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknow"
		}

		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()

		p := c.Request.URL.Path
		start := time.Now()

		entry1 := logger.WithFields(logrus.Fields{
			"hostname":  hostname,
			"clientIP":  clientIP,
			"method":    c.Request.Method,
			"path":      p,
			"referer":   referer,
			"userAgent": clientUserAgent,
		})
		entry1.Info("[REQ_BEG] " + reqId)

		c.Next()

		statusCode := c.Writer.Status()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		entry2 := logger.WithFields(logrus.Fields{
			"statusCode": statusCode,
			"latency":    latency, // time to process
			"dataLength": dataLength,
		})

		if len(c.Errors) > 0 {
			entry2.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			if statusCode >= http.StatusInternalServerError {
				entry2.Error("[REQ_END] " + reqId)
			} else if statusCode >= http.StatusBadRequest {
				entry2.Warn("[REQ_END] " + reqId)
			} else {
				entry2.Info("[REQ_END] " + reqId)
			}
		}
	}
}
