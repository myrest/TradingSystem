package services

import "TradingSystem/src/common"

var systemsettings common.SystemSettings

func init() {
	systemsettings = common.GetEnvironmentSetting()
	//初始化Aduit，要在systemsettings讀取完後
	initAduit()
}
