package telegram

import (
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/pkg/logging"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"os"
)

// TODO переделать на DB структуру
// TODO переделать на SQLX модуль
type User struct {
	chatid   int64
	status   string
	username string
}

var bot *tgbotapi.BotAPI

func SendMessageToTelegramWithLogError(errorText string) {
	logger := logging.GetLogger()
	logger.Println("Start SendMessageToTelegramWithLogError")
	defer logger.Println("End SendMessageToTelegramWithLogError")
	logger.Error(errorText)
	err := SendMessage(errorText)
	if err != nil {
		logger.Errorf("Не удалось отправить сообщение в телеграм: error: %v", err)
		// tODO оправить на почту
	}
}

// отправить сообщение в телеграм юзерам из DB
func SendMessage(messageText string) error {

	logger := logging.GetLoggerWithSeviceName("telegram")
	logger.Info("Bot.Send:>Start")
	defer logger.Info("Bot.Send:>End")

	// select DB - получить user-ов
	users, err := GetDbUsers()
	if err != nil {
		return errors.Wrapf(err, "failed GetDbUsers")
	}

	if len(users) > 0 {
		for chatID, user := range users {
			logger.Info(chatID, user)

			for {

				if len(messageText) == 0 {
					break
				}

				if len(messageText) > 4096 {

					messageTextRune := []rune(messageText[:4096])

					logger.Debugf("len(messageText):%d", len(messageText))
					logger.Debugf("len([]rune(messageText)):%d", len([]rune(messageText)))

					msg := tgbotapi.NewMessage(chatID, string(messageTextRune[:len(messageTextRune)-1]))
					msg.ParseMode = "HTML"
					_, err = bot.Send(msg)
					if err != nil {
						return errors.Wrapf(err, "failed bot.Send(%v)", msg)
					}

					logger.Debugf("messageText до обрезки, len=%d:\n%s", len(messageText), messageText)

					messageText = string([]rune(messageText)[len(messageTextRune)-1:])

					logger.Debugf("messageText обрезано:\n%s", string(messageTextRune[:len(messageTextRune)-1]))
					logger.Debugf("messageText после обрезки, len=%d:\n%s", len(messageText), messageText)

				} else {

					logger.Debugf("messageText, len=%d:\n%s", len(messageText), messageText)

					msg := tgbotapi.NewMessage(chatID, messageText)
					msg.ParseMode = "HTML"
					_, err = bot.Send(msg)
					if err != nil {
						return errors.Wrapf(err, "failed bot.Send(%v)", msg)
					}
					break
				}

			}

		}
	} else {
		return errors.New("Bot.Send.Message:>Users in db is none")
	}
	logger.Debugf(
		"Bot.Send.Message:>%v", messageText)

	return nil
}

// проверка, что файл существует
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// создание базы для телеграм
func CreateDB() error {

	logger := logging.GetLoggerWithSeviceName("telegram")
	logger.Info("CreateDB:>Start")
	defer logger.Info("CreateDB:>End")

	logger.Info("CreateDB:>Creating telegram.db...")

	db, err := sql.Open("sqlite3", "telegram.db")
	if err != nil {
		errorText := fmt.Sprintf(`failed sql.Open("sqlite3", "telegram.db"), error: %v`, err)
		logger.Fatalf(errorText)
		return errors.New(errorText)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Infof("failed db.Close(), error: %v", err)
		}
	}(db)

	logger.Info("CreateDB:>telegram.db created")

	createUsersTableSQL := `CREATE TABLE USERS (
		"chatid" integer NOT NULL PRIMARY KEY UNIQUE,		
		"status" TEXT,
        "username" TEXT
	  );` // SQL Statement for Create Table

	logger.Info("CreateDB:>Create USERS table...")
	statement, err := db.Prepare(createUsersTableSQL) // Prepare SQL Statement
	if err != nil {
		errorText := fmt.Sprintf("db.Prepare(createUsersTableSQL), error: %v, createUsersTableSQL:\n%s", err, createUsersTableSQL)
		logger.Fatal(errorText)
		return errors.New(errorText)
	}
	result, err := statement.Exec()
	if err != nil {
		return errors.Wrapf(err, "failed in statement.Exec()")
	} else {
		logger.Debugf("statement.Exec() result %v", result)
	}

	logger.Info("CreateDB:>USERS created")

	return nil
}

// добавить пользователся в db
func InsertDbUsers(user *User, db *sql.DB) error {

	logger := logging.GetLoggerWithSeviceName("telegram")
	logger.Info("InsertDbUsers:>Start")
	defer logger.Info("InsertDbUsers:>End")
	// выполняем Insert
	result, err := db.Exec("insert into USERS (chatid, status, username) values ($1, $2, $3)",
		user.chatid, user.status, user.username)
	if err != nil {
		return errors.Wrapf(err, `failed db.Exec("insert into USERS (chatid, status, username) values (%d, %s, %s)"`, user.chatid, user.status, user.username)
	}

	logger.Debug("InsertDbUsers:>ID добавленного объекта")
	logger.Debug(result.LastInsertId()) // id последнего добавленного объекта
	logger.Debug("InsertDbUsers:>Количество добавленных строк")
	logger.Debug(result.RowsAffected()) // количество добавленных строк

	return nil

}

// обновить статус пользователя
func UpdateStatusDbUsers(user *User, db *sql.DB) error {

	logger := logging.GetLoggerWithSeviceName("telegram")
	logger.Info("UpdateStatusDbUsers:>Start")
	defer logger.Info("UpdateStatusDbUsers:>End")

	// выполняем Insert
	result, err := db.Exec("update USERS set status = $1 where chatid = $2 ",
		user.status, user.chatid)
	if err != nil {
		return errors.Wrapf(err, `failed db.Exec("update USERS set status = %s where chatid = %d`, user.status, user.chatid)
	}

	logger.Debug("UpdateStatusDbUsers:>Количество обновленных строк")
	logger.Debug(result.RowsAffected()) // количество обновленных строк

	return nil
}

//получить users из db
func GetDbUsers() (map[int64]User, error) {

	logger := logging.GetLoggerWithSeviceName("telegram")
	logger.Info("GetDbUsers:>Start")
	defer logger.Info("GetDbUsers:>End")

	db, err := sql.Open("sqlite3", "telegram.db")
	if err != nil {
		return nil, errors.Wrap(err, `failed sql.Open("sqlite3", "telegram.db")`)
	} else {
		logger.Info("GetDbUsers.DB:>telegram.db open")
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Infof("failed db.Close(), error: %v", err)
		}
	}(db)

	// выполняем select
	rows, err := db.Query("select * from USERS")
	if err != nil {
		return nil, errors.Wrap(err, `db.Query("select * from USERS")`)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Infof("failed rows.Close(), error: %v", err)
		}
	}(rows)

	users := make(map[int64]User)

	for rows.Next() {
		c := User{}
		err := rows.Scan(&c.chatid, &c.status, &c.username)
		if err != nil {
			return nil, errors.Wrapf(err, "rows.Scan(&c.chatid=%d, &c.status=%s, &c.username=%s)", c.chatid, c.status, c.username)
		}
		users[c.chatid] = c //считываем все users, кому будем выполнять отправку message
	}
	return users, nil
}

//запуск сервиса telegram bot
func BotStart() {

	logger := logging.GetLoggerWithSeviceName("telegram")
	logger.Info("BotStart:>Start")
	defer logger.Info("BotStart:>End")

	if Exists("telegram.db") != true {
		logger.Info("BotStart.DB:>telegram.db not exist")
		err := CreateDB()
		if err != nil {
			logger.Fatalf("BotStart.DB.Error:>Not create telegram.db, %v", err)
		}
	} else {
		logger.Info("BotStart.DB:>telegram.db exist")
	}

	db, err := sql.Open("sqlite3", "telegram.db")
	if err != nil {
		logger.Fatalf("BotStart.DB:>telegram.db not open, %v", err)
		return
	} else {
		logger.Info("BotStart.DB:>telegram.db open")
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Infof("failed db.Close(), error: %v", err)
		}
	}(db)

	err = tgbotapi.SetLogger(logger)
	if err != nil {
		logger.Fatalf("BotStart.tgbotapi.Setlogger.Print:>%v", err)
	}

	cfg := config.GetConfig()
	if cfg.TELEGRAM.Debug == 0 {
		bot.Debug = false
	} else {
		bot.Debug = true
	}

	logger.Infof("BotStart:>Authorized on account %v", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		logger.Fatalf("BotStart.GetUpdatesChan.Error:>%v", err)
		// TODO!! если ошибка при обработке то отправить письмо или любое другое оповещение, что телеграм сломался
		// Обязательно, потому что может быть потеря заказов
		return
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		// получить пользователей из базы для рассылки
		users, err := GetDbUsers()
		if err != nil {
			logger.Infof("BotStart.GetDbUsers.Error:>%v", err)
			continue
		}

		chatid := update.Message.Chat.ID

		// TODO сделать админов в телеграме для удаления и добавления пользователей
		if user, ok := users[chatid]; ok { // проверяем юзера полученный из Message
			// user найден, то проверяем Status и обновляем
			if user.status == "0" {
				user.status = "1" // если пришло сообщение от пользователя, то значит он активный и делаем его активным со статусом = 1

				// обновляем DB если он поменялся
				err := UpdateStatusDbUsers(&user, db)
				if err != nil {
					logger.Infof("BotStart.UpdateStatusDbUsers.Error:>%v", err)
					continue
				}
			}

			logger.Infof("BotStart:>user: %v ", user)
		} else { // юзер не найден, поэтому мы его добавляем в базу
			u := &User{
				chatid:   chatid,
				status:   "1",
				username: update.Message.From.UserName,
			}

			users[chatid] = *u
			err := InsertDbUsers(u, db)
			if err != nil {
				logger.Infof("BotStart.InsertDbUsers.Error:>%v", err)
				continue
			}

			logger.Infof("BotStart:>User add: %d", users[chatid])
		}

		//---
		//сюда добавить обработчик телеграм запросов
		logger.Infof("BotStart.Message:>[%s] %s", update.Message.From.UserName, update.Message.Text)

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вам включены оповещения")
		msg.ReplyToMessageID = update.Message.MessageID

		//оправляем ответ-эхо
		_, err = bot.Send(msg)
		if err != nil {
			logger.Infof("Bot.Send.Error:>%v", err)
			continue
		}
		logger.Info(users)
	}
}

func init() {

	logger := logging.GetLoggerWithSeviceName("telegram")
	logger.Info("init telegram bot:>Start")
	defer logger.Info("init telegram bot:>End")

	var err error
	cfg := config.GetConfig()

	bot, err = tgbotapi.NewBotAPI(cfg.TELEGRAM.BotToken)
	if err != nil {
		logger.Panicf("failed init, error: %v", err)
	}

}
