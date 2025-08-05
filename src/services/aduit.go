package services

import (
	"TradingSystem/src/common"
	"context"
	"log"
	"sync"

	"cloud.google.com/go/logging"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var (
	logger     *logging.Logger
	client     *logging.Client
	loggerMux  sync.RWMutex
	loggerOnce sync.Once
)

const (
	AuditCollectionName = "Audit"
	LoggerName          = "AuditLog"
)

type AuditLogType string
type EventNameType string

const (
	customerEvent AuditLogType = "CustomerEvent"
	systemEvent   AuditLogType = "SystemEvent"

	EventNameLogin      EventNameType = "LoginEvent"
	EventNameSystemInit EventNameType = "SystemInit"
	PlaceOrder          EventNameType = "PlaceOrder"
	WeeklyReport        EventNameType = "WeeklyReport"
)

type CustomerEventLog struct {
	EventName  EventNameType
	CustomerID string
	Message    interface{}
}

type SystemEventLog struct {
	EventName  EventNameType
	LogType    AuditLogType
	CustomerID string
	Message    interface{}
}

type customerEventLogDB struct {
	CustomerEventLog
	LogType   AuditLogType
	RemoteIP  string
	EventTime string
}

type systemEventLogDB struct {
	SystemEventLog
	LogType   AuditLogType
	RemoteIP  string
	EventTime string
}

func initAudit() {
	loggerOnce.Do(func() {
		settings := common.GetFirebaseSetting()
		var sa option.ClientOption
		ctx := context.Background()

		if common.IsFileExists(settings.FireBaseKeyFullPath) {
			sa = option.WithCredentialsFile(settings.FireBaseKeyFullPath)
		} else {
			creds, err := common.GetSecret(ctx)
			if err != nil {
				log.Fatalf("failed to access secret version: %v", err)
			}
			sa = option.WithCredentialsJSON([]byte(creds))
		}

		var err error
		client, err = logging.NewClient(ctx, settings.ProjectID, sa)
		if err != nil {
			log.Fatalf("Failed to create logging client: %v", err)
		}

		loggerMux.Lock()
		logger = client.Logger(LoggerName)
		loggerMux.Unlock()
	})
}

func getLogger() *logging.Logger {
	loggerMux.RLock()
	defer loggerMux.RUnlock()
	return logger
}

func FlushLogging() {
	if l := getLogger(); l != nil {
		l.Flush()
	}
}

func CloseLogging() error {
	loggerMux.Lock()
	defer loggerMux.Unlock()

	if client != nil {
		return client.Close()
	}
	return nil
}

func (e CustomerEventLog) SendWithoutIP() {
	e.sendMessage()
}

func (e CustomerEventLog) Send(c *gin.Context) {
	e.sendMessage(c)
}

func (e CustomerEventLog) sendMessage(cs ...*gin.Context) {
	go func() {
		if e.Message == nil {
			return
		}

		currentLogger := getLogger()
		if currentLogger == nil {
			log.Printf("Logger not initialized, falling back to standard log: %+v", e)
			return
		}

		client_ip := "0.0.0.0"
		if len(cs) > 0 {
			client_ip = cs[0].ClientIP()
		}

		entry := logging.Entry{
			Severity: logging.Notice,
			Payload: customerEventLogDB{
				CustomerEventLog: e,
				EventTime:        common.GetUtcTimeNow(),
				RemoteIP:         client_ip,
				LogType:          customerEvent,
			},
		}

		currentLogger.Log(entry)
	}()
}

func (e SystemEventLog) Send() {
	go func() {
		if e.Message == nil {
			return
		}

		currentLogger := getLogger()
		if currentLogger == nil {
			log.Printf("Logger not initialized, falling back to standard log: %+v", e)
			return
		}

		entry := logging.Entry{
			Severity: logging.Notice,
			Payload: systemEventLogDB{
				SystemEventLog: e,
				EventTime:      common.GetUtcTimeNow(),
				LogType:        systemEvent, // 修正：使用正確的 systemEvent
			},
		}

		currentLogger.Log(entry)
	}()
}
