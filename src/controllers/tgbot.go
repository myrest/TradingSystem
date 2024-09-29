package controllers

import (
	"TradingSystem/src/common"
	"TradingSystem/src/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var tgbot *tgbotapi.BotAPI

type tgCommandType struct {
	IsNeedExtraInfo bool
	Function        func(context.Context, int64, string)
}

func runHelpCmd(c context.Context, chatID int64, cmdList map[string]tgCommandType) {
	cmdFunc, isexist := cmdList["help"]
	if isexist {
		cmdFunc.Function(c, chatID, "")
	} else {
		//顯示預設的Help
		commandStart(c, chatID, "")
	}
}

// Commnad開始
// 第一層命令
var tgBotCommandRoot = map[string]tgCommandType{
	"/set": {
		IsNeedExtraInfo: true,
		Function:        commandGroupSet,
	},
	"/unset": {
		IsNeedExtraInfo: true,
		Function:        commandGroupUnSet,
	},
	"/start": {
		Function: commandStart,
	},
	"help": {
		Function: commandStart,
	},
	"/help": {
		Function: commandStart,
	},
	"/list": {
		Function: commandList,
	},
}

func commandStart(c context.Context, chatID int64, param string) {
	msg := "目前支援以下命令\n/set ID [您專屬的識別碼]	啟用您的訊息通知。"
	msg += "\n/unset ID [您專屬的識別碼]	停用您的訊息通知。"
	msg += "\n/list	列出己綁定帳號。"
	msg += "\n/help or /start	列出目前支援的命令。"
	resp := tgbotapi.NewMessage(chatID, msg)
	tgbot.Send(resp)
}

func commandGroupSet(c context.Context, chatID int64, param string) { //第二層的頭
	runTGCommand(c, chatID, param, tgBotCommandSet)
}

func commandGroupUnSet(c context.Context, chatID int64, param string) { //第二層的頭
	runTGCommand(c, chatID, param, tgBotCommandUnSet)
}

// 列出己綁定帳號
func commandList(c context.Context, chatID int64, param string) {
	customers, err := services.GetCustomerByTgChatID(c, chatID)
	if err != nil {
		resp := tgbotapi.NewMessage(chatID, "無法取得您的資料，請稍候再試。")
		tgbot.Send(resp)
		return
	}

	var rtnarr []string
	for _, customer := range *customers {
		if common.IsEmail(customer.Email) {
			rtnarr = append(rtnarr, fmt.Sprintf("%s\t識別碼：%s", customer.Email, customer.TgIdentifyKey))
		} else {
			rtnarr = append(rtnarr, fmt.Sprintf("%s\t識別碼：%s", customer.Name, customer.TgIdentifyKey))
		}
	}

	if len(rtnarr) == 0 {
		resp := tgbotapi.NewMessage(chatID, "您還沒有綁定任何帳號唷。")
		tgbot.Send(resp)
	}

	rtn := "己綁定帳號如下：\n" + strings.Join(rtnarr[:], "\n") //轉成字串

	resp := tgbotapi.NewMessage(chatID, rtn)
	tgbot.Send(resp)
}

// 第二層 Set 命令
var tgBotCommandSet = map[string]tgCommandType{
	"id": {
		IsNeedExtraInfo: true,
		Function:        commandSetID,
	},
	"help": {
		IsNeedExtraInfo: true,
		Function:        commandSetHelp,
	},
}

func commandSetHelp(c context.Context, chatID int64, param string) {
	resp := tgbotapi.NewMessage(chatID, "請使用\n/set ID [您專屬的識別碼]\n來啟用您的訊息通知。")
	tgbot.Send(resp)

}

func commandSetID(c context.Context, chatID int64, param string) {
	//先檢查有沒有傳入tg token
	tgIdentifyKey := param
	customer, err := services.GetCustomerByTgIdentifyKey(c, tgIdentifyKey)
	if err != nil || customer == nil {
		if param == "" {
			resp := tgbotapi.NewMessage(chatID, "請輸入的您的專屬識別碼。")
			tgbot.Send(resp)
		} else {
			resp := tgbotapi.NewMessage(chatID, "您輸入的專屬識別碼有誤。")
			tgbot.Send(resp)
		}
		return
	}

	customer.TgChatID = chatID
	err = services.UpdateCustomer(c, customer)
	if err != nil {
		resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("無法更新您的資料。[%s]", err.Error()))
		tgbot.Send(resp)
		return
	}

	isEmail := common.IsEmail(customer.Email)
	resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("您的資料已與帳號：[%s](%s)綁定完成。", customer.Email, customer.Name))
	if !isEmail {
		resp = tgbotapi.NewMessage(chatID, fmt.Sprintf("您的資料已與子帳號：%s 綁定完成。", customer.Name))
	}
	tgbot.Send(resp)
}

// 第二層 UnSet 命令
var tgBotCommandUnSet = map[string]tgCommandType{
	"id": {
		IsNeedExtraInfo: true,
		Function:        commandUnSetID,
	},
}

func commandUnSetID(c context.Context, chatID int64, param string) {
	tgIdentifyKey := param
	customer, err := services.GetCustomerByTgIdentifyKey(c, tgIdentifyKey)
	if err != nil || customer == nil {
		resp := tgbotapi.NewMessage(chatID, "您輸入的專屬識別碼有誤。")
		tgbot.Send(resp)
		return
	}

	customer.TgChatID = 0
	err = services.UpdateCustomer(c, customer)
	if err != nil {
		resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("無法更新您的資料。[%s]", err.Error()))
		tgbot.Send(resp)
		return
	}

	isEmail := common.IsEmail(customer.Email)
	resp := tgbotapi.NewMessage(chatID, fmt.Sprintf("您的資料已與帳號：[%s](%s)解除綁定。", customer.Email, customer.Name))
	if !isEmail {
		resp = tgbotapi.NewMessage(chatID, fmt.Sprintf("您的資料已與子帳號：%s 解除綁定。", customer.Name))
	}
	tgbot.Send(resp)
}

//Commnad結束

func runTGCommand(c context.Context, chatID int64, cmd string, cmdList map[string]tgCommandType) {
	parts := strings.Fields(cmd) // 將輸入拆分
	if len(parts) == 0 {
		runHelpCmd(c, chatID, cmdList)
		return
	}

	commandStr := strings.ToLower(parts[0]) // 取得命令並轉小寫
	var parameter string
	if len(parts) > 1 {
		parameter = strings.Join(parts[1:], " ") //取得剩餘字串
	}

	//如果找不到命令
	cmdFunc, isexist := cmdList[commandStr]
	if !isexist {
		runHelpCmd(c, chatID, cmdList)
		return
	}

	//如果有額外訊息需要處理
	if cmdFunc.IsNeedExtraInfo {
		//執行該Function
		cmdFunc.Function(c, chatID, parameter)
		return
	}

	if parameter == "" {
		cmdFunc.Function(c, chatID, "")
		return
	}

	runHelpCmd(c, chatID, cmdList)
}

func init() {
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

func handleUpdate(c *gin.Context, update tgbotapi.Update) {
	//username := update.Message.From.UserName
	//if username == "" {
	//	username = fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
	//}

	chatID := update.Message.Chat.ID
	if chatID == 0 {
		return //沒有找到ChatID，直接忽略
	}

	fmt.Println(update.Message.Text)
	//從第一層執行起
	runTGCommand(c, chatID, update.Message.Text, tgBotCommandRoot)
}
