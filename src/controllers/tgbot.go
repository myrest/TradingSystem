package controllers

import (
	"TradingSystem/src/common"
	"TradingSystem/src/services"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var tgbot *tgbotapi.BotAPI
var checkedChatID map[int64]struct{}

func init() {
	checkedChatID = make(map[int64]struct{})
	settings := common.GetEnvironmentSetting()
	go func() {
		bot, err := tgbotapi.NewBotAPI(settings.TgToken)
		if err != nil {
			log.Panic(err)
		}
		tgbot = bot
	}()
}

func TGbot(c *gin.Context) {
	var update tgbotapi.Update
	// 从请求体中解析 Telegram 更新
	if err := c.ShouldBindJSON(&update); err != nil {
		log.Println("Could not decode incoming update:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 处理 Telegram 消息
	if update.Message != nil {
		handleUpdate(c, update)
	}

	// 响应 Telegram，告知接收到更新
	c.Status(http.StatusOK)
}

func int64InMap(set map[int64]struct{}, value int64) bool {
	_, exists := set[value]
	return exists
}

func handleUpdate(c *gin.Context, update tgbotapi.Update) {
	username := update.Message.From.UserName
	if username == "" {
		username = fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
	}

	chatID := update.Message.Chat.ID
	if chatID == 0 {
		return //沒有找到ChatID，直接忽略
	}

	customer, err := services.GetCustomerByTGChatID(c, chatID)
	if err != nil {
		resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("無法取得您的資料。[%s]", err.Error()))
		tgbot.Send(resp)
		return
	}

	if customer != nil {
		//有啟用過
		handleExistingCustomer(chatID, username)
	} else {
		//未啟用過
		handleNewCustomer(update, c, chatID, username)
	}
}

func handleNewCustomer(update tgbotapi.Update, c *gin.Context, chatID int64, username string) {
	prefix := "/set id "
	if strings.HasPrefix(strings.ToLower(update.Message.Text), prefix) {
		tgIdentifyKey := strings.TrimPrefix(update.Message.Text, prefix)
		customer, err := services.GetCustomerByTgIdentifyKey(c, tgIdentifyKey)
		if err != nil || customer == nil {
			resp := tgbotapi.NewMessage(chatID, "無法取得您的資料。")
			tgbot.Send(resp)
		} else {
			customer.TgChatID = chatID
			err := services.UpdateCustomer(c, customer)
			if err != nil {
				resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("無法更新您的資料。[%s]", err.Error()))
				tgbot.Send(resp)
			} else {
				resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("您的資料已與帳號：[%s](%s)綁定完成。", customer.Email, customer.Name))
				tgbot.Send(resp)
			}
		}
	} else {
		resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("[%s]您好。請使用\n/set ID [您專屬的識別碼]\n來啟用您的訊息通知。", username))
		tgbot.Send(resp)
	}
}

func handleExistingCustomer(chatID int64, username string) {
	if int64InMap(checkedChatID, chatID) {
		resp := tgbotapi.NewMessage(chatID, getRandomEmoticon())
		tgbot.Send(resp)
	} else {
		checkedChatID[chatID] = struct{}{}
		resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("[%s]您好。您的設定己完成，請耐心等待訊息的通知。", username))
		tgbot.Send(resp)
	}
}

func getRandomEmoticon() string {
	emoticons := []string{
		"請耐心等待相關事件發生，自統會自動傳送訊息給您。",
		"請耐心等待相關事件發生，自統會自動傳送訊息給您。",
		"請耐心等待相關事件發生，自統會自動傳送訊息給您。",
		"請耐心等待相關事件發生，自統會自動傳送訊息給您。",
		"請耐心等待相關事件發生，自統會自動傳送訊息給您。",
		"請耐心等待相關事件發生，自統會自動傳送訊息給您。",
		"請耐心等待相關事件發生，自統會自動傳送訊息給您。",
		"̿̿ ̿̿ ̿̿ ̿'̿’\\̵͇̿̿\\з=( ͠° ͟ʖ ͡°)=ε/̵͇̿̿/‘̿̿ ̿ ̿ ̿ ̿ ̿",
		"(′゜ω。‵)",
		"₍₍ ◝('ω'◝) ⁾⁾ ₍₍ (◟'ω')◟ ⁾⁾",
		"(｡･㉨･｡)",
		"( ͡° ͜ʖ ͡°)",
		"(๑´ڡ`๑)",
		"(●´ω｀●)ゞ",
		"♥(´∀` )人",
		"(ﾉ◕ヮ◕)ﾉ*:･ﾟ✧",
		"ヾ(●゜▽゜●)♡",
		"(๑•̀ㅂ•́)و✧",
		"✧◝(⁰▿⁰)◜✧",
		"(♡˙︶˙♡)",
	}

	// 生成一个随机的索引值
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := seededRand.Intn(len(emoticons))
	// 返回随机的表情符号
	return emoticons[randomIndex]
}
