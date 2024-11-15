package common

import (
	"context"
	"fmt"

	"google.golang.org/api/iterator"
)

func GetDBSystemSettings(ctx context.Context) (SystemSettings, error) {
	var rtn SystemSettings

	client := GetFirestoreClient()

	// 查詢 SystemSettings 集合，限制結果為 1
	docs, err := client.Collection("SystemSettings").Limit(1).Documents(ctx).GetAll()
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
	iter := client.Collection("SystemSettings").Where("Env", "==", settings.Env).Limit(1).Documents(ctx)
	defer iter.Stop() // 確保迭代器關閉

	doc, err := iter.Next()
	if err == iterator.Done {
		// 如果找不到文檔，則創建一個新文檔
		_, _, err := client.Collection("SystemSettings").Add(ctx, settings)
		if err != nil {
			return err
		}
	} else {
		// 找到文檔，更新其值
		_, err := doc.Ref.Set(ctx, settings)
		if err != nil {
			return err
		}
		ApplySystemSettings(settings)
	}

	return nil
}
