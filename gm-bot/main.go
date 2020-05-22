package main 

import (
	"github.com/bwmarrin/discordgo"
	"github.com/willdrug/viper" // using the fork since the original gets no updates
	"github.com/willdrug/dice-string-parser" 
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"strings"
)

func init() {
	var err error
	viper.SetConfigName("gm_main") // name of config file (without extension)
	viper.AddConfigPath(".") // local
	//viper.AddConfigPath("/config/") // for future Dockerfile
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize config storage: %#v", err))
	}
	// Mostly to remember what should be in config
	// viper.SetDefault("api_key", nil)
}

func main() {
	// check necessary config values
	api_key := viper.GetString("api_key") 
	if api_key == "" {
		fmt.Println("Empty API key. Add api_key into configuration file.")
	}

	// init bot
	gm, err := discordgo.New("Bot " + api_key)
	if err != nil {
		fmt.Println(err)
	}

	gm.AddHandler(ready)
	gm.AddHandler(messageCreate)

	err = gm.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	fmt.Println("Airhorn is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	gm.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, "Testing")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// check if the message is "!airhorn"
	if strings.HasPrefix(m.Content, "#quit") {
		s.Close()
	}
	if strings.HasPrefix(m.Content, "#ping") {
		s.ChannelMessageSend(m.ChannelID, "Echo!")
	}
	if strings.HasPrefix(m.Content, "#roll") {
		rp, err := diceparser.Parse(m.Content[5:])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		var b, e bool
		b = false
		e = false
		for _, v := range rp.Rolls() {
			for _, q := range v.Rolls() {
				if q > v.Dice() {
					e = true
				}
				if q < 0 {
					b = true
				}
			}
		}
		ret := fmt.Sprintf("Rolling %s\nRolls: [", rp.Dstring())
		for _, v:= range rp.Rolls() {
			ret += fmt.Sprintf("\n\t%s: Kept: %v; Discarded: %v; Bonus: %d; Malus: %d; Explode: %d; Botch: %d; Limit: %d; Total: %d", v.Rstring(), v.Rolls(), v.Discarded(), v.Bonus(), v.Malus(), v.Explode(), v.Botch(), v.Lim(), v.Total())
		}
		ret += fmt.Sprintf("\n]", )
		if e {
			ret += "\n*EXPLOSION!*"
		}
		if b {
			ret += "\n*BOTCH!*"
		}
		ret += fmt.Sprintf("\nTotal: %d", rp.Total())
		s.ChannelMessageSend(m.ChannelID, ret)
	}
}