package services

import "TradingSystem/src/common"

var systemsettings common.SystemSettings

func init() {
	systemsettings = common.GetEnvironmentSetting()
	//初始化Audit，要在systemsettings讀取完後
	initAudit()
}
