package main

import (
	"encoding/json"
"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
	
	"strings"
	
	"my-telegram-task-planner-bot/mongo"
	"go.mongodb.org/mongo-driver/bson"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var timerKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("10 min"),
		tgbotapi.NewKeyboardButton("15 min"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("20 min"),
		tgbotapi.NewKeyboardButton("30 min"),
	),
)

type taskTime struct{
	Task string
	Timer string
}
	

var yesKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("yes"),
		tgbotapi.NewKeyboardButton("no"),
	),
)

// Config represents the structure of the config.json file
type Config struct {
	Token string `json:"token"`
}

func readConfig() Config {
	file, err := os.Open("config.json")
	if err != nil {
		log.Panic(err)
	}

	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panic(err)
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		log.Panic(err)
	}

	return config
}

func mongoConnect() mongo.Collection{
	c:=mongo.NewClient("","","localhost",27017)
	c.Connection()
	collection:=c.NewCollection("test","mycollection",true)
	return collection
}

func createTaskButtons(tt []taskTime)( button tgbotapi.ReplyKeyboardMarkup) {
	var exitFlag int
	var coveredFlag int
	a:=tgbotapi.NewKeyboardButton("")
	b:=tgbotapi.NewKeyboardButtonRow(a)	
	c:=tgbotapi.NewKeyboardButtonRow(a)	
	d:=tgbotapi.NewKeyboardButtonRow(a)	
	e:=tgbotapi.NewKeyboardButtonRow(a)	
	f:=tgbotapi.NewKeyboardButtonRow(a)	
	taskNo:=0
	for _,v:= range tt{
		if v.Task!=""{
			switch taskNo{
			case 0:
				
				a:=tgbotapi.NewKeyboardButton(v.Task)
				b=tgbotapi.NewKeyboardButtonRow(a)
				taskNo++	
			case 1:
				
				a:=tgbotapi.NewKeyboardButton(v.Task)
				c=tgbotapi.NewKeyboardButtonRow(a)
				taskNo++	
			case 2:
				
				a:=tgbotapi.NewKeyboardButton(v.Task)
				d=tgbotapi.NewKeyboardButtonRow(a)	
				taskNo++
			case 3:
				
				a:=tgbotapi.NewKeyboardButton(v.Task)
				e=tgbotapi.NewKeyboardButtonRow(a)
				taskNo++	
			case 4:
				
				a:=tgbotapi.NewKeyboardButton(v.Task)
				f=tgbotapi.NewKeyboardButtonRow(a)
				button=	tgbotapi.NewReplyKeyboard(b,c,d,e,f)
				coveredFlag=4	
			 }
		}
		exitFlag=1
		}
		if exitFlag==1 && coveredFlag!=4{
			if taskNo==1{
			
			button=	tgbotapi.NewReplyKeyboard(b)
			return
		}else if taskNo==2{
		
			button=	tgbotapi.NewReplyKeyboard(b,c)
			return
		}else if taskNo==3{
			
			button=	tgbotapi.NewReplyKeyboard(b,c,d)
			return
		}else if taskNo==4{
			
			button=	tgbotapi.NewReplyKeyboard(b,c,d,e)
			return

		}
	}
return button
}


func getTimer(tt []taskTime,m string) string{
	for _,t:= range tt{
	switch t.Task{
	case m:
		return t.Timer
		}	
	}
	return ""
}

func findTaskTime(c mongo.Collection,d []taskTime) []taskTime{
	err:=c.Find(bson.M{},bson.M{},&d)
	if err!=nil{
		log.Println(err)
	}
	fmt.Println(d)
	return d
}

func deleteFinishedTask(c mongo.Collection,m string){
	err:=c.DeleteOne(bson.M{"task":m})
	if err!=nil{
		fmt.Println(err)
	}
}

func setSleepTimer(timer string){
	timeVal:=strings.Split(timer,"")
	getTime:=timeVal[0]
	switch getTime{
	case "30":
		time.Sleep(30 * time.Minute)
	case "20":
		time.Sleep(20 * time.Minute)
	case "15":
		time.Sleep(15 * time.Minute)
	case "10":
		time.Sleep(10 * time.Minute)

	}
}

func botConfig() (bot *tgbotapi.BotAPI,err error){
	config := readConfig()
	bot, err = tgbotapi.NewBotAPI(config.Token)
	
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)
    return bot,err
}
func getUpdateChannel(bot *tgbotapi.BotAPI)(updates tgbotapi.UpdatesChannel,err error){
	updateConfig := tgbotapi.NewUpdate(5)
	updateConfig.Timeout = 60
	updates, err = bot.GetUpdatesChan(updateConfig)
	return updates,err
}

func main() {
	var todo,timer string
	var insertFlag int
	var startFlag=0
    var tasktime []taskTime
	var commandFlag string
	var taskNo=0
	bot,err:=botConfig()
	if err != nil {
		log.Panic(err)
	}
	updates ,err:=getUpdateChannel(bot)
	if err != nil {
		log.Panic(err)
	}
	collection:=mongoConnect()
	
	for update := range updates {

		 if update.Message == nil {
			continue
	    }
		
		if commandFlag=="todo"{
		todo=update.Message.Text
			fmt.Printf("Todo:%s",todo)
			commandFlag="off"
			insertFlag=1
		}

		if commandFlag == "time"{
			if update.Message.Text!="setime"{
				timer=update.Message.Text
				fmt.Printf("Time:%s",timer)
				commandFlag="off"
				insertFlag=2
			}
		}

		if insertFlag==2{
			err:=collection.InsertOne(bson.M{"task":todo,"timer":timer})
			insertFlag=0
			if err!=nil{
				log.Println(err)
			}
			taskNo++
		}

		if taskNo==5{	
			tasktime=findTaskTime(collection,tasktime) 
			taskNo=0
		}

	
	updateMsg:=update.Message.Text
	
		switch update.Message.Text {
		
		case "todo":
			commandFlag="todo"
		case "setime":
			commandFlag="time"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,"Please set the timer for your task")
			msg.ReplyMarkup = timerKeyboard
			bot.Send(msg)
		case "start":
			fmt.Println(tasktime)	
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,"Which task do you want to start first?")
			msg.ReplyMarkup =   createTaskButtons(tasktime)
			bot.Send(msg)
			startFlag=1
		case updateMsg:
			if startFlag==1{
			
				timer:=getTimer(tasktime,updateMsg)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,"Please start your Task: "+updateMsg+" !!!\nStarting Timer for "+timer)
				bot.Send(msg)
				findTaskTime(collection,tasktime)
				deleteFinishedTask(collection,updateMsg)
				setSleepTimer(timer)
				msg=tgbotapi.NewMessage(update.Message.Chat.ID,timer+" is over! Did you complete your task?")
				msg.ReplyMarkup = yesKeyboard
				bot.Send(msg)
				startFlag=2			
			}

			if startFlag==2{
				if updateMsg=="yes"{
					msg := tgbotapi.NewMessage(update.Message.Chat.ID,"Good Job! Keep going!!")
						bot.Send(msg)
						tasktime=findTaskTime(collection,tasktime)
					if tasktime!=nil{
						msg = tgbotapi.NewMessage(update.Message.Chat.ID,"Which task do you want to start next?")
						msg.ReplyMarkup =   createTaskButtons(tasktime)
						bot.Send(msg)
						startFlag=1
					}
				}
					if updateMsg=="no"{
						msg := tgbotapi.NewMessage(update.Message.Chat.ID,"Oops! Waiting.....  Did you complete your task?")
						bot.Send(msg)
					}			
				
			}
			
		
			
		}
	
	
	}
	

}


