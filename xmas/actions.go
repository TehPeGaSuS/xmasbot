package xmas

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lrstanley/girc"
	"github.com/ringsaturn/tzf"
)

const helpMsg = "Commands: '%sxmas <location>', '%stime <location>', '%snext', '%sprevious', '%sremaining', '%shelp', '%ssource'"

var tzFinder tzf.F

func init() {
	f, err := tzf.NewDefaultFinder()
	if err != nil {
		panic(err)
	}
	tzFinder = f
}

func (bot *Settings) addTriggers() {
	irc := bot.irc

	//Log Notices
	irc.Handlers.Add(girc.NOTICE, func(c *girc.Client, e girc.Event) {
		bot.logger.Info("[NOTICE] " + e.Last())
	})

	//Trigger for !source
	irc.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		if !strings.HasPrefix(normalize(e.Last()), bot.Prefix+"source") {
			return
		}
		c.Cmd.ReplyTo(e, "Source code: https://github.com/TehPeGaSuS/xmasbot")
	})

	//Trigger for !help
	irc.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		content := normalize(e.Last())
		if !strings.HasPrefix(content, bot.Prefix+"help") && content != bot.Prefix+"xmas" {
			return
		}
		bot.logger.Info("Querying help...")
		c.Cmd.ReplyTo(e, fmt.Sprintf(helpMsg, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix))
	})

	//Trigger for !next
	irc.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		if !strings.HasPrefix(normalize(e.Last()), bot.Prefix+"next") {
			return
		}
		bot.logger.Info("Querying next...")
		dur := time.Minute * time.Duration(bot.next.Offset*60)
		if now().UTC().Add(dur).After(bot.target) {
			c.Cmd.ReplyTo(e, "No more next, Christmas is here AoE")
			return
		}
		hdur := humanDur(bot.target.Sub(now().UTC().Add(dur)))
		hdur = bot.col(hdur)
		var next = bot.col("Next Merry Christmas") + " in "
		max := c.MaxEventLength()
		max -= len(next)
		max -= len(hdur)
		max -= 4
		c.Cmd.ReplyTo(e, next+hdur+" in "+bot.next.Format(max, bot.Colors))
	})

	//Trigger for !previous
	irc.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		content := normalize(e.Last())
		if !strings.HasPrefix(content, bot.Prefix+"prev") && !strings.HasPrefix(content, bot.Prefix+"last") {
			return
		}
		bot.logger.Info("Querying previous...")
		dur := time.Minute * time.Duration(bot.previous.Offset*60)
		hdur := humanDur(now().UTC().Add(dur).Sub(bot.target))
		if bot.previous.Offset == -12 {
			hdur = humanDur(now().UTC().Add(dur).Sub(bot.target.AddDate(-1, 0, 0)))
		}
		hdur = bot.col(hdur)
		var prev = bot.col("Previous Merry Christmas") + " was "
		max := c.MaxEventLength()
		max -= len(prev)
		max -= len(hdur)
		max -= 8
		c.Cmd.ReplyTo(e, prev+hdur+" ago in "+bot.previous.Format(max, bot.Colors))
	})

	//Trigger for !remaining
	irc.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		if !strings.HasPrefix(normalize(e.Last()), bot.Prefix+"remaining") {
			return
		}
		bot.logger.Info("Querying remaining...")
		plural := "s"
		if bot.remaining == 1 {
			plural = ""
		}
		pct := ((float64(len(bot.zones)) - float64(bot.remaining)) / float64(len(bot.zones)) * 100)
		c.Cmd.ReplyTo(e, fmt.Sprintf("%d timezone%s remaining. %.2f%% are in the Christmas day", bot.remaining, plural, pct))
	})

	//Trigger for time in location
	irc.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		content := normalize(e.Last())
		if !strings.HasPrefix(content, bot.Prefix+"time ") {
			return
		}
		bot.logger.Info("Querying time...")
		arg := content[len(bot.Prefix)+len("time")+1:]
		msg, err := timeInTZ(arg)
		if err == nil {
			c.Cmd.ReplyTo(e, msg)
			return
		}
		if errors.Is(err, ErrTZParser) {
			bot.logger.Warn("Query: " + err.Error())
			c.Cmd.ReplyTo(e, err.Error())
			return
		}

		result, err := bot.time(arg)
		if err == errNoZone || err == errNoPlace {
			bot.logger.Warn("Query error: " + err.Error())
			c.Cmd.ReplyTo(e, err.Error())
			return
		}
		if err != nil {
			bot.logger.Warn("Query error: " + err.Error())
			c.Cmd.ReplyTo(e, "Some error occurred!")
			return
		}
		c.Cmd.ReplyTo(e, result)
	})

	//Trigger for UTC time
	irc.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		if normalize(e.Last()) != bot.Prefix+"time" {
			return
		}
		bot.logger.Info("Querying time...")
		result := "Time is " + now().UTC().Format("Mon Jan 2 15:04:05 -0700 MST 2006")
		c.Cmd.ReplyTo(e, result)
	})

	//Trigger for merry christmas in location
	irc.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		content := normalize(e.Last())
		if !strings.HasPrefix(content, bot.Prefix+"xmas ") {
			return
		}
		arg := content[len(bot.Prefix)+len("xmas")+1:]
		msg, err := bot.newChristmasInTZ(arg)
		if err == nil {
			c.Cmd.ReplyTo(e, msg)
			return
		}
		if errors.Is(err, ErrTZParser) {
			bot.logger.Warn("Query: " + err.Error())
			c.Cmd.ReplyTo(e, err.Error())
			return
		}
		result, err := bot.newChristmas(arg)
		if err == errNoZone || err == errNoPlace {
			bot.logger.Warn("Query error: " + err.Error())
			c.Cmd.ReplyTo(e, err.Error())
			return
		}
		if err != nil {
			bot.logger.Warn("Query error: " + err.Error())
			c.Cmd.ReplyTo(e, "Some error occurred!")
			return
		}

		c.Cmd.ReplyTo(e, result)
	})
}

var (
	errNoZone  = errors.New("couldn't get timezone for that location")
	errNoPlace = errors.New("couldn't find that place")
)

func lookupZone(lat, lon float64) (*time.Location, error) {
	tzid := tzFinder.GetTimezoneName(lon, lat)
	if tzid == "" {
		return nil, errNoZone
	}
	zone, err := time.LoadLocation(tzid)
	if err != nil {
		return nil, errNoZone
	}
	return zone, nil
}

func (bot *Settings) time(location string) (string, error) {
	bot.logger.Info("Querying location: " + location)

	res, err := NominatimFetcher(bot.Email, bot.Nominatim, location)
	if err != nil {
		bot.logger.Warn("Nominatim error: " + err.Error())
		return "", err
	}
	if len(res) == 0 {
		return "", errNoPlace
	}
	zone, err := lookupZone(res[0].Lat, res[0].Lon)
	if err != nil {
		return "", err
	}
	address := res[0].DisplayName
	msg := fmt.Sprintf("Time in %s is %s", address, now().In(zone).Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
	return msg, nil
}

func (bot *Settings) newChristmas(location string) (string, error) {
	bot.logger.Info("Querying location: " + location)
	res, err := NominatimFetcher(bot.Email, bot.Nominatim, location)
	if err != nil {
		bot.logger.Warn("Nominatim error: " + err.Error())
		return "", err
	}
	if len(res) == 0 {
		return "", errNoPlace
	}
	zone, err := lookupZone(res[0].Lat, res[0].Lon)
	if err != nil {
		return "", err
	}
	offset := zoneOffset(bot.target, zone)
	address := res[0].DisplayName
	if now().UTC().Add(offset).Before(bot.target) {
		hdur := humanDur(bot.target.Sub(now().UTC().Add(offset)))
		const newChristmasFutureMsg = "Merry Christmas in %s will happen in %s"
		return fmt.Sprintf(newChristmasFutureMsg, address, hdur), nil
	}
	hdur := humanDur(now().UTC().Add(offset).Sub(bot.target))
	const newChristmasPastMsg = "Merry Christmas in %s happened %s ago"
	return fmt.Sprintf(newChristmasPastMsg, address, hdur), nil
}

func (bot *Settings) newChristmasInTZ(tzAbbr string) (msg string, err error) {
	tzAbbr = strings.ToUpper(tzAbbr)
	var offset int
	var ok bool
	if offset, ok = tzAbbrs[tzAbbr]; !ok {
		offset, err = parseUTC(tzAbbr)
		if err != nil {
			return "", err
		}
	}

	offsetdur := time.Duration(offset) * time.Second
	t := now()
	if t.UTC().Add(offsetdur).Before(bot.target) {
		hdur := humanDur(bot.target.Sub(t.UTC().Add(offsetdur)))
		const newChristmasFutureMsg = "Merry Christmas in %s will happen in %s"
		return fmt.Sprintf(newChristmasFutureMsg, tzAbbr, hdur), nil
	}

	hdur := humanDur(t.UTC().Add(offsetdur).Sub(bot.target))
	const newChristmasPastMsg = "Merry Christmas in %s happened %s ago"
	return fmt.Sprintf(newChristmasPastMsg, tzAbbr, hdur), nil
}

func timeInTZ(tzAbbr string) (msg string, err error) {
	tzAbbr = strings.ToUpper(tzAbbr)
	var offset int
	var ok bool
	if offset, ok = tzAbbrs[tzAbbr]; !ok {
		offset, err = parseUTC(tzAbbr)
		if err != nil {
			return "", err
		}
	}
	t := now()
	msg = fmt.Sprintf("Time in %s is %s", tzAbbr, t.In(time.FixedZone(tzAbbr, offset)).Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
	return msg, nil
}
