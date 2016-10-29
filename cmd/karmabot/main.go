package main

import (
	"flag"

	"github.com/kamaln7/karmabot"
	"github.com/kamaln7/karmabot/database"
	karmabotui "github.com/kamaln7/karmabot/ui"
	"github.com/kamaln7/karmabot/ui/blankui"
	"github.com/kamaln7/karmabot/ui/webui"

	"github.com/aybabtme/log"
	"github.com/nlopes/slack"
)

// cli flags
var (
	token            = flag.String("token", "", "slack RTM token")
	dbpath           = flag.String("db", "./db.sqlite3", "path to sqlite database")
	maxpoints        = flag.Int("maxpoints", 6, "the maximum amount of points that users can give/take at once")
	leaderboardlimit = flag.Int("leaderboardlimit", 10, "the default amount of users to list in the leaderboard")
	debug            = flag.Bool("debug", false, "set debug mode")
	webuitotp        = flag.String("webui.totp", "", "totp key")
	webuipath        = flag.String("webui.path", "", "path to web UI files")
	webuilistenaddr  = flag.String("webui.listenaddr", "", "address to listen and serve the web ui on")
	webuiurl         = flag.String("webui.url", "", "url address for accessing the web ui")
	motivate         = flag.Bool("motivate", true, "toggle motivate.im support")
)

func main() {
	// startup

	ll := log.KV("version", karmabot.Version)
	ll.Info("starting karmabot")

	// config
	flag.Parse()

	// database

	db, err := database.New(&database.Config{
		Path: *dbpath,
	})

	if err != nil {
		ll.KV("path", *dbpath).Err(err).Fatal("could not open sqlite db")
	}

	// slack

	if *token == "" {
		ll.Fatal("please pass the slack RTM token (see `karmabot -h` for help)")
	}

	//TODO: figure out a way to fix this
	//our current logging library does not implement
	//log.Logger
	//slack.SetLogger(*ll)
	sc := &karmabot.Slack{
		Bot: slack.New(*token),
	}
	sc.Bot.SetDebug(*debug)
	sc.RTM = sc.Bot.NewRTM()

	go sc.RTM.ManageConnection()

	// karmabot

	var ui karmabotui.Provider
	if *webuipath != "" && *webuilistenaddr != "" {
		ui, err = webui.New(&webui.Config{
			ListenAddr:       *webuilistenaddr,
			URL:              *webuiurl,
			FilesPath:        *webuipath,
			TOTP:             *webuitotp,
			LeaderboardLimit: *leaderboardlimit,
			Log:              ll.KV("provider", "webui"),
			Debug:            *debug,
			DB:               db,
		})

		if err != nil {
			ll.Err(err).Fatal("could not initialize web ui")
		}
	} else {
		ui = blankui.New()
	}
	go ui.Listen()

	bot := karmabot.New(&karmabot.Config{
		Slack:            sc,
		UI:               ui,
		Debug:            *debug,
		MaxPoints:        *maxpoints,
		LeaderboardLimit: *leaderboardlimit,
		Log:              ll,
		DB:               db,
	})

	bot.Listen()
}