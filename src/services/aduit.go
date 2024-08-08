package services

import (
	"TradingSystem/src/common"
	"context"
	"log"

	"cloud.google.com/go/logging"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var logger *logging.Logger

const (
	AduitColliectionName = "Aduit"
	LoggerName           = "AduitLog"
)

type AduitLogType string
type EventNameType string

const (
	customerEvent AduitLogType = "CustomerEvent"
	systemEvent   AduitLogType = "SystemEvent"

	EventNameLogin EventNameType = "LoginEvent"
)

type CustomerEventLog struct {
	EventName  EventNameType
	CustomerID string
	Message    interface{}
}

type SystemEventLog struct {
	EventName  EventNameType
	LogType    AduitLogType
	CustomerID string
	Message    interface{}
}

type customerEventLogDB struct {
	CustomerEventLog
	LogType   AduitLogType
	RemoteIP  string
	EventTime string
}

type systemEventLogDB struct {
	SystemEventLog
	LogType   AduitLogType
	RemoteIP  string
	EventTime string
}

func init() {
	settings := common.GetEnvironmentSetting()
	var sa option.ClientOption
	ctx := context.Background()
	if common.IsFileExists(settings.FireBaseKeyFullPath) {
		sa = option.WithCredentialsFile(settings.FireBaseKeyFullPath)
	} else {
		creds, err := getSecret(ctx, "projects/635522974118/secrets/GOOGLE_APPLICATION_CREDENTIALS/versions/latest")
		if err != nil {
			log.Fatalf("failed to access secret version: %v", err)
		}
		sa = option.WithCredentialsJSON([]byte(creds))
	}

	client, err := logging.NewClient(ctx, settings.ProjectID, sa)
	if err != nil {
		log.Fatalf("Failed to create logging client: %v", err)
	}

	// 获取一个 Logger
	go func() {
		logger = client.Logger(LoggerName) //啟動超過30秒，時間花太久。
	}()
}

func FlushLogging() {
	logger.Flush()
}

func (e CustomerEventLog) Send(c *gin.Context) {
	go func() {
		if e.Message != nil {
			entry := logging.Entry{
				//Timestamp: time.Now(),
				Severity: logging.Notice,
				Payload: customerEventLogDB{
					CustomerEventLog: e,
					EventTime:        common.GetUtcTimeNow(),
					RemoteIP:         c.ClientIP(),
					LogType:          customerEvent,
				},
			}
			if logger != nil {
				logger.Log(entry)
			} else {
				log.Printf("%v", entry)
			}
		}
	}()
}

func (e SystemEventLog) Send() {
	go func() {
		if e.Message != nil {
			entry := logging.Entry{
				//Timestamp: time.Now(),
				Severity: logging.Notice,
				Payload: systemEventLogDB{
					SystemEventLog: e,
					EventTime:      common.GetUtcTimeNow(),
					LogType:        customerEvent,
				},
			}
			if logger != nil {
				logger.Log(entry)
			} else {
				log.Printf("%v", entry)
			}
		}
	}()
}
