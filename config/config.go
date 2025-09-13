package config

import (
	"m3u8dweb/db"
	"m3u8dweb/models"
)

// 初始化默认设置（如果数据库中没有设置）
func InitDefaultSettings() {
	var settings models.SystemSettings
	err := db.GetData("settings", "system", &settings)
	
	// 如果没有找到设置或发生错误，使用默认设置
	if err != nil || settings.SaveLocation == "" {
		defaultSettings := models.DefaultSettings()
		db.SaveData("settings", "system", defaultSettings)
	}
}
