// Merry Christmas IRC party bot
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/TehPeGaSuS/xmasbot/xmas"
	"github.com/badoux/checkmail"
	"github.com/fatih/color"
	log "gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/yaml.v3"
)

const usage = `
Merry Christmas IRC party bot
Announces merry christmas as they happen in each timezone

CMD Options:
[mandatory]
-channels	comma separated list of channels eg. "#test, #test2"
		channel key can be specifed after ":" e.g #channelname:channelkey
-nick		irc nick
-email		nominatim email

[optional]
-password	irc server password
-saslnick	sasl username
-saslpass	sasl password
-server		irc server (default: irc.libera.chat:6697)
-prefix		command prefix (default: !)
-nossl		disable ssl for irc
-nocheck	disable ssl verification (e.g. self signed ssl)
-nominatim	nominatim server (default: http://nominatim.openstreetmap.org)
-nolimit	disable flood kick protection
-colors		enable irc colors
-bind		bind to host address
-debug		debug irc traffic
-yaml		yaml config file

`
const SET_NOMINATIM_SERVER = "https://nominatim.openstreetmap.org"
const SET_LIBERA_SERVER = "irc.libera.chat:6697"
const SET_PREFIX = "!"

func main() {
	var channels xmas.Channels

	// Mandatory
	flag.Var(&channels, "channels", "comma separated list of channels")
	nick := flag.String("nick", "", "irc nick")
	email := flag.String("email", "", "nominatim email")
	// Optional
	password := flag.String("password", "", "irc server password")
	saslNick := flag.String("saslnick", "", "sasl username")
	saslPass := flag.String("saslpass", "", "sasl password")
	server := flag.String("server", SET_LIBERA_SERVER, "irc server")
	prefix := flag.String("prefix", SET_PREFIX, "command prefix")
	nossl := flag.Bool("nossl", false, "disable ssl for irc")
	nocheck := flag.Bool("nocheck", false, "disable ssl verification (e.g. self signed ssl)")
	nominatim := flag.String("nominatim", SET_NOMINATIM_SERVER, "nominatim server")
	nolimit := flag.Bool("nolimit", false, "disable limit bot replies.")
	colors := flag.Bool("colors", false, "enable irc colors")
	debug := flag.Bool("debug", false, "debug irc traffic")
	bind := flag.String("bind", "", "bind to host address")
	configYAML := flag.String("yaml", "", "use yaml settings file")

	green := color.New(color.FgGreen)
	flag.Usage = func() {
		green.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	c := config{
		{
			Nick:      *nick,
			Channels:  channels,
			Server:    *server,
			NoSSL:     *nossl,
			NoCheck:   *nocheck,
			Password:  *password,
			SaslNick:  *saslNick,
			SaslPass:  *saslPass,
			Prefix:    *prefix,
			Email:     *email,
			Nominatim: *nominatim,
			NoLimit:   *nolimit,
			Colors:    *colors,
			Debug:     *debug,
			Bind:      *bind,
		},
	}

	red := color.New(color.FgHiRed)

	if *configYAML != "" {
		data, err := os.ReadFile(*configYAML)
		if err != nil {
			red.Fprintln(os.Stderr, "yaml file: ", err)
			os.Exit(1)
		}
		c = config{}
		err = yaml.Unmarshal(data, &c)
		if err != nil {
			red.Fprintln(os.Stderr, "yaml file: ", err)
			os.Exit(1)
		}
		for i := range c {
			if c[i].Nominatim == "" {
				c[i].Nominatim = SET_NOMINATIM_SERVER
			}
			if c[i].Server == "" {
				c[i].Server = SET_LIBERA_SERVER
			}
			if c[i].Prefix == "" {
				c[i].Prefix = SET_PREFIX
			}
		}
	}

	err := check(c)
	if err != nil {
		red.Fprintln(os.Stderr, err)
		if *configYAML == "" {
			flag.Usage()
		}
		os.Exit(1)
	}
	var bots []*xmas.Settings
	for _, c := range c {
		bot, err := xmas.New(
			&xmas.Settings{
				Nick:      c.Nick,
				Channels:  c.Channels,
				Server:    c.Server,
				SSL:       !c.NoSSL,
				NoCheck:   c.NoCheck,
				Bind:      c.Bind,
				Password:  c.Password,
				SaslNick:  c.SaslNick,
				SaslPass:  c.SaslPass,
				Prefix:    c.Prefix,
				Email:     c.Email,
				Nominatim: c.Nominatim,
				Limit:     !c.NoLimit,
				Colors:    c.Colors,
				Debug:     c.Debug,
			},
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, "BINDHOST:", err)
			os.Exit(1)
		}
		bots = append(bots, bot)
	}

	for i, bot := range bots {
		if c[i].Debug {
			bot.LogLvl(log.LvlDebug)
		} else {
			bot.LogLvl(log.LvlInfo)
		}
		go bot.Start()
	}
	select {}
}

type config []struct {
	Nick      string
	Channels  []string
	Server    string
	NoSSL     bool
	NoCheck   bool
	Password  string
	SaslNick  string
	SaslPass  string
	Prefix    string
	Email     string
	Nominatim string
	NoLimit   bool
	Colors    bool
	Debug     bool
	Bind      string
}

func check(c config) error {
	if len(c) == 0 {
		return fmt.Errorf("empty or misconfigured yaml")
	}

	for _, c := range c {

		// Check mandatory inputs
		if len(c.Channels) == 0 {
			return fmt.Errorf("error: no channels defined")
		}
		channelReg := regexp.MustCompile(`^([#&][^\x07\x2C\s]{0,200})$`)
		for _, ch := range c.Channels {
			if !channelReg.MatchString(ch) {
				return fmt.Errorf("error: invalid channel name: %s", ch)
			}
		}
		if c.Nick == "" {
			return fmt.Errorf("error: no nick defined")
		}
		if len(c.Nick) > 16 {
			return fmt.Errorf("error: nick too long")
		}
		if c.Email == "" {
			return fmt.Errorf("error: no email provided")
		}
		if err := checkmail.ValidateFormat(c.Email); err != nil {
			return fmt.Errorf("error: invalid email address")
		}
		// Check optional inputs
		if c.Server == "" {
			return fmt.Errorf("error: no irc server defined")
		}
		serverReg := regexp.MustCompile(`^\S+:\d+$`)
		if !serverReg.MatchString(c.Server) {
			return fmt.Errorf("error: invalid irc server address")
		}
		if c.Prefix == "" {
			return fmt.Errorf("error: no command prefix defined")
		}
		prefixReg := regexp.MustCompile(`^\W+$`)
		if !prefixReg.MatchString(c.Prefix) {
			return fmt.Errorf("error: prefix must be non-alphanumeric")
		}
		if c.Nominatim == "" {
			return fmt.Errorf("error: no nominatim server provided")
		}
	}
	return nil
}
