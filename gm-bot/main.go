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

func env(key string) (string, error) {
	for _, e := range os.Environ() {
        pair := strings.SplitN(e, "=", 2)
        if pair[0] == key {
        	return pair[1], nil
        }
    }
    return "", fmt.Errorf("Key %s not present in env", key)
}

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
	viper.SetDefault("api_key", "API_KEY") // holds environment key to get the actual api key
}

func main() {
	var err error
	// check necessary config values
	api_key := viper.GetString("api_key") 
	if api_key == "" {
		fmt.Println("Empty API key key. Add api_key into configuration file.")
		return
	}
	api_key, err = env(api_key)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// init bot
	gm, err := discordgo.New("Bot " + api_key)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	gm.AddHandler(ready)
	gm.AddHandler(messageCreate)

	err = gm.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err.Error())
		return
	}

	fmt.Println("Bot's running")
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
	var ret string
	var err error
	// check if the message is "!airhorn"
	if strings.HasPrefix(m.Content, "#quit") {
		s.Close()
	}
	if strings.HasPrefix(m.Content, "#ping") {
		ret = "Echo"
	}
	if strings.HasPrefix(m.Content, "#roll") {
		ret, err = Roll(m.Content[5:])
		if err != nil {
			ret = err.Error()
		}
	}
	s.ChannelMessageSend(m.ChannelID, ret)
}

func Roll(rstring) string, err {
	rp, err := diceparser.Parse(rstring)
	if err != nil {
		return "", err
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
	return ret, nil
}