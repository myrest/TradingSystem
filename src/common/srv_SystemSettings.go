package common

import (
	"context"
	"fmt"
)

func GetDBSystemSettings(ctx context.Context) (SystemSettings, error) {
	var rtn SystemSettings

	client := GetFirestoreClient()

	// 查詢 SystemSettings 集合，限制結果為 1
	docs, err := client.Collection("SystemSettings").Where("Env", "==", systemSettings.Env.String()).Limit(1).Documents(ctx).GetAll()
	if err != nil {
		return rtn, err
	}

	// 檢查是否有資料
	if len(docs) == 0 {
		//return nil, fmt.Errorf("查無系統設定資料。")
		fmt.Printf("查無系統設定資料。")
		return rtn, nil
	}

	docs[0].DataTo(&rtn)

	return rtn, nil
}

func SaveDBSystemSettings(ctx context.Context, settings SystemSettings) error {
	client := GetFirestoreClient()
	doc, err := client.Collection("SystemSettings").Where("Env", "==", settings.Env.String()).Limit(1).Documents(ctx).GetAll()

	if err != nil {
		return err
	}

	if len(doc) == 0 {
		// 如果找不到文檔，則創建一個新文檔
		_, _, err := client.Collection("SystemSettings").Add(ctx, settings)
		if err != nil {
			return err
		}
	} else {
		// 找到文檔，更新其值
		_, err := doc[0].Ref.Set(ctx, settings)
		if err != nil {
			return err
		}
		ApplySystemSettings(settings)
	}

	return nil
}
