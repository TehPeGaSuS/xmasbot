package xmas

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lrstanley/girc"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Settings for the bot
type Settings struct {
	Nick      string
	Channels  []string
	Server    string
	SSL       bool
	NoCheck   bool
	Bind      string
	Password  string
	SaslNick  string
	SaslPass  string
	Prefix    string
	Email     string
	Nominatim string
	Limit     bool
	Colors    bool
	Debug     bool
	irc       *girc.Client
	dialer    *net.Dialer
	logger    log.Logger
	extra
}

type extra struct {
	zones     TZS
	previous  TZ
	next      TZ
	remaining int
	first     bool
	target    time.Time
	channels  []string
	joined    chan struct{}
}

// New creates a new bot
func New(s *Settings) (*Settings, error) {
	host, port, err := net.SplitHostPort(s.Server)
	if err != nil {
		return nil, err
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	s.dialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if s.Bind != "" {
		localAddr, err := net.ResolveIPAddr("ip", s.Bind)
		if err != nil {
			return nil, err
		}
		s.dialer.LocalAddr = &net.TCPAddr{IP: localAddr.IP}
	}

	var sasl girc.SASLMech
	if s.SaslPass != "" {
		sasl = &girc.SASLPlain{User: s.SaslNick, Pass: s.SaslPass}
	}

	s.joined = make(chan struct{})
	for _, ch := range s.Channels {
		name, _, _ := strings.Cut(ch, ":")
		s.channels = append(s.channels, name)
	}

	s.logger = log.New()

	ircConfig := girc.Config{
		Server:     host,
		Port:       portNum,
		Nick:       s.Nick,
		User:       s.Nick,
		Name:       s.Nick,
		ServerPass: s.Password,
		SASL:       sasl,
		SSL:        s.SSL,
		TLSConfig:  &tls.Config{InsecureSkipVerify: s.NoCheck},
		AllowFlood: !s.Limit,
	}
	if s.Debug {
		ircConfig.Debug = &debugWriter{logger: s.logger}
	}
	s.irc = girc.New(ircConfig)

	s.irc.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		for _, ch := range s.Channels {
			name, key, hasKey := strings.Cut(ch, ":")
			if hasKey {
				c.Cmd.JoinKey(name, key)
			} else {
				c.Cmd.Join(name)
			}
		}
		close(s.joined)
	})

	s.addTriggers()

	return s, nil
}

// debugWriter routes girc's raw protocol debug output through our logger
type debugWriter struct{ logger log.Logger }

func (d *debugWriter) Write(p []byte) (int, error) {
	d.logger.Debug(strings.TrimRight(string(p), "\r\n"))
	return len(p), nil
}

// LogLvl sets the log level
func (bot *Settings) LogLvl(Lvl log.Lvl) {
	bot.logger.SetHandler(log.LvlFilterHandler(Lvl, log.StderrHandler))
}

// Start starts the bot
func (bot *Settings) Start() {
	bot.target = target
	bot.logger.Info("Starting the bot...")

	go bot.ircControl()

	<-bot.joined
	// Need to wait a bit for channels to actually finish joining
	time.Sleep(time.Second * 5)
	bot.logger.Info("Got start...")

	if err := bot.decodeZones(Zones); err != nil {
		bot.logger.Crit("Decode zones error: " + err.Error())
		return
	}
	for {
		bot.loopTimeZones()
		var zonesFinishedMsg = bot.col("That's it") + ", Christmas " +
			bot.col("%d") + " is here " +
			bot.col("Anywhere on Earth")
		for _, ch := range bot.channels {
			bot.irc.Cmd.Message(ch, fmt.Sprintf(zonesFinishedMsg, bot.target.Year()))
		}
		bot.logger.Info("All zones finished...")
		bot.target = bot.target.AddDate(1, 0, 0)
		bot.logger.Info(fmt.Sprintf("Wrapping the target date around to %d", bot.target.Year()))
	}
}

func (bot *Settings) decodeZones(z []byte) error {
	if err := json.Unmarshal(z, &bot.zones); err != nil {
		return err
	}
	sort.Sort(sort.Reverse(bot.zones))
	return nil
}

const reconnectInterval = time.Second * 30

func (bot *Settings) ircControl() {
	for {
		if err := bot.irc.DialerConnect(bot.dialer); err != nil {
			bot.logger.Warn("connect error: " + err.Error())
		}
		bot.logger.Info("Reconnecting...")
		time.Sleep(reconnectInterval)
	}
}

func (bot *Settings) loopTimeZones() {
	zones := bot.zones
	irc := bot.irc
	for i := 0; i < len(zones); i++ {
		dur := time.Minute * time.Duration(zones[i].Offset*60)
		bot.next = zones[i]
		if i == 0 {
			bot.previous = zones[len(zones)-1]
		} else {
			bot.previous = zones[i-1]
		}
		bot.remaining = len(zones) - i
		if now().UTC().Add(dur).Before(bot.target) {
			time.Sleep(time.Second * 2)
			bot.logger.Info(fmt.Sprintf("Zone pending: %.2f", zones[i].Offset))
			hdur := humanDur(bot.target.Sub(now().UTC().Add(dur)))
			hdur = bot.col(hdur)
			next := bot.col("Next Merry Christmas") + " in "
			if i == 0 && !(now().Month() == time.December && now().Day() < 26) {
				next = bot.col("First Merry Christmas") + " in "
			}
			if i == len(zones)-1 {
				next = bot.col("Final Merry Christmas") + " in "
			}
			help := fmt.Sprintf(helpMsg, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix)
			for _, ch := range bot.channels {
				max := irc.MaxEventLength()
				max -= len(next)
				max -= len(hdur)
				max -= 4
				if !bot.first {
					irc.Cmd.Message(ch, next+hdur+" in "+zones[i].Format(max, bot.Colors))
					irc.Cmd.Message(ch, help)
					bot.first = true
				} else {
					irc.Cmd.Message(ch, next+hdur+". "+
						fmt.Sprintf("See %snext or %shelp.", bot.Prefix, bot.Prefix))
				}
			}
			//Wait till Target in Timezone
			timer := NewTimer(bot.target.Sub(now().UTC().Add(dur)))
			<-timer.C
			timer.Stop()
			var happy = bot.col("Merry Christmas") + " in "
			for _, ch := range bot.channels {
				max := irc.MaxEventLength()
				max -= len(happy)
				irc.Cmd.Message(ch, happy+zones[i].Format(max, bot.Colors))
			}
			bot.logger.Info(fmt.Sprintf("Announcing zone: %.2f", zones[i].Offset))
		}
	}
}

// https://modern.ircdocs.horse/formatting.html
func (bot *Settings) col(s string) string {
	if bot.Colors {
		s = "\x02\x0303" + s + "\x0f"
	}
	return s
}
