package services

import "TradingSystem/src/common"

var systemsettings common.SystemSettings

func init() {
	systemsettings = common.GetEnvironmentSetting()
}
