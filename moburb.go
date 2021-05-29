package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/mileusna/crontab"
	log "github.com/sirupsen/logrus"
	"github.com/yanzay/tbot/v2"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type application struct {
	client *tbot.Client
}

var (
	app   application
	bot   *tbot.Server
	token string
	ctab  *crontab.Crontab
)

func init() {
	//token = os.Getenv("TELEGRAM_TOKEN")
	ctab = crontab.New()

	log.SetReportCaller(true)
	var filename string = "moburb.log"
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		// Cannot open log file. Logging to stderr
		fmt.Println(err)
	} else {
		log.SetOutput(f)
	}

	formatter := &log.TextFormatter{
		TimestampFormat:        "02-01-2006 15:04:05", // the "time" field configuratiom
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	}

	log.SetFormatter(formatter)

	//    log.Trace("Something very low level.")
	//    log.Debug("Useful debugging information.")
	//    log.Info("Something noteworthy happened!")
	//    log.Warn("You should probably take a look at this.")
	//    log.Error("Something failed but I'm not quitting.")
	//    log.Fatal("Bye.") Calls os.Exit(1) after logging
	//    log.Panic("I'm bailing.") Calls panic() after logging
	log.SetLevel(log.InfoLevel)
}

func main() {
	bot = tbot.New(token)
	app.client = bot.Client()

	bot.HandleMessage("/start", app.startHandler)
	bot.HandleMessage("/stop", app.stopHandler)
	bot.HandleMessage("/status", app.statusHandler)
	bot.HandleMessage("/check", app.checkOnceHandler)

	err := bot.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func (a *application) startHandler(m *tbot.Message) {
	if a.checkUser(m) {
		return
	}

	ctab.Clear()
	ctab.MustAddJob("5 12,19 * * 1-6", checkMobUrb, a.client, m, true)
	log.Info("Job set: mon-fry, 12:05, 19:05")
	//ctab.MustAddJob("* * * * *", checkMobUrb, a.client, m, true)
}

func (a *application) stopHandler(m *tbot.Message) {
	if a.checkUser(m) {
		return
	}
	ctab.Clear()
	msg := "Am oprit crontab"
	_, err := a.client.SendMessage(m.Chat.ID, msg)
	if err != nil {
		log.Error(err)
	}
	log.Info("Am oprit crontab")
	return
}

func (a *application) statusHandler(m *tbot.Message) {
	if a.checkUser(m) {
		return
	}
	msg := "I'm up"
	_, err := a.client.SendMessage(m.Chat.ID, msg)
	if err != nil {
		log.Error(err)
	}
	return
}

func (a *application) checkOnceHandler(m *tbot.Message) {
	if a.checkUser(m) {
		return
	}
	checkMobUrb(a.client, m, false)
	return
}

func checkMobUrb(bot *tbot.Client, m *tbot.Message, job bool) {
	if job {
		log.Debug("Started cron job")
	}
	// Verific blocatoare
	url := "https://www.mobilitateurbana4.ro/registratura-online/"
	response, err := http.Get(url)
	if err != nil {
		log.Errorf("Err get no %v: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Errorf("Response status NOK on get %v")
	}

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Errorf("Error loading HTTP response to doc.: %v", err)
	}
	// Close response body if requests went ok
	response.Body.Close()
	document.Find("script").Each(func(i int, selection *goquery.Selection) {
		departments, _ := regexp.MatchString("departments", selection.Text())

		msg := ""

		if departments {
			match, _ := regexp.MatchString("blocatoare", selection.Text())

			if match {
				if !job {
					msg = "Nu sunt blocatoare"
					_, err = bot.SendMessage(m.Chat.ID, msg)
				}
				log.Debugf("Job: %v - Nu sunt blocatoare", job)
			} else {
				msg = "Am gasit blocatoare"
				_, err = bot.SendMessage(m.Chat.ID, msg)
				log.Debug(msg)

			}
			if err != nil {
				log.Error(err)
			}
		}
	})
	// Verificare blocatoare -- pana aici

	// Verific disponibiliate  loc de parcare
	//url := "https://resedinta.mobilitateurbana4.ro/ajax/check-available"
	//method := "POST"
	//
	//payload := &bytes.Buffer{}
	//writer := multipart.NewWriter(payload)
	//
	//_ = writer.WriteField("id", "4915") // 9386       4915
	//err := writer.Close()
	//if err != nil {
	//	log.Error(err)
	//}
	//
	//client := &http.Client{}
	//
	//req, err := http.NewRequest(method, url, payload)
	//
	//if err != nil {
	//	log.Error(err)
	//}
	//req.Header.Set("Content-Type", writer.FormDataContentType())
	//res, err := client.Do(req)
	//if err != nil {
	//	log.Error(err)
	//}
	//defer res.Body.Close()
	//body, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	log.Error(err)
	//}
	//
	//avlbl := false
	//
	//log.Debugf("Raspuns moburb: %v", string(body))
	//if string(body) == "1" {
	//	avlbl = true
	//}
	//msg := ""
	//
	//if avlbl {
	//	msg = "Am gasit loc de parcare"
	//	_, err = bot.SendMessage(m.Chat.ID, msg)
	//	log.Debug(msg)
	//} else {
	//	if !job {
	//		msg = "Nu am gasit loc de parcare"
	//		_, err = bot.SendMessage(m.Chat.ID, msg)
	//	}
	//	log.Debugf("Job: %v - Nu am gasit loc de parcare", job)
	//}
	//if err != nil {
	//	log.Error(err)
	//}
	// Verific disponibiliate  loc de parcare -- pana aici

}

func (a *application) checkUser(m *tbot.Message) (notAllowed bool) {
	userID, err := strconv.Atoi(m.Chat.ID)

	if err != nil {
		log.Fatal(err)
	}

	if userID != 44872081 {
		log.Warning("Alien user doing stuff")
		_, err = a.client.SendMessage(m.Chat.ID, "You don't have access")
		if err != nil {
			log.Error("Err sending message: ", err)
		}

		err = a.client.LeaveChat(m.Chat.ID)
		if err != nil {
			log.Warning("Cannot leave private chat")
		}
		notAllowed = true
	}
	return notAllowed
}
