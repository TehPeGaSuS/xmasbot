package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/TehPeGaSuS/xmasbot/xmas"
	"github.com/lrstanley/girc"
)

func main() {
	client := girc.New(girc.Config{
		Server: "testnet.ergo.chat",
		// Server: "irc.libera.chat",
		Port: 6697,
		Nick: "happyhappy2025v2",
		User: "happyhappy2025v2",
		Name: "happyhappy2025v2",
		SSL:  true,
		Out:  log.Writer(),
	})

	var zones xmas.TZS
	err := json.Unmarshal(xmas.Zones, &zones)
	if err != nil {
		log.Fatal(err)
	}

	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		c.Cmd.Join("##xmas00")
	})

	client.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		if e.Last() != "!test" {
			return
		}
		for _, z := range zones {
			const pre = "\x02\x0302Next Merry Christmaas\x0f in \x02\x030213 seconds 323 milliseconds\x0f in "
			c.Cmd.ReplyTo(e, pre+z.Format(c.MaxEventLength()-len(pre), true))
			time.Sleep(time.Second)
			c.Cmd.ReplyTo(e, "**************************")
			time.Sleep(time.Second * 3)
		}
	})

	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
}
