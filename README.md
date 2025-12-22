# xmasbot
Merry Christmas bot based on https://github.com/ugjka/newyearsbot
----
# XmasBot

## 2025 here we come

Merry Christmas IRC party bot

Posts Merry Christmas for each timezone when they happen
## Bot's commands

- `!next` upcoming Merry Christmas
- `!previous` previous Merry Mhristmas
- `!remaining` number of remaining timezones
- `!xmas <location>` get Merry Christmas status for location
- `!time <location>` get the current time in a location
- `!time` UTC
- `!help` show help

The command prefix `!` can be changed using the -prefix flag

## Pro-tip

- make sure your system's time is synchronized with NTP

## Installation

### Using make

You need to have make, go, go-tools

Build with `make`

Install with `make install`

Uninstall with `make uninstall`

Clean up with `make clean`

## Usage

```
[bots@codebay-server xmasbot]$ ./xmasbot -h

Merry Christmas IRC party bot
Announces merry christmas as they happen in each timezone

CMD Options:
[mandatory]
-channels       comma separated list of channels eg. "#test, #test2"
                channel key can be specifed after ":" e.g #channelname:channelkey
-nick           irc nick
-email          nominatim email

[optional]
-password       irc server password
-saslnick       sasl username
-saslpass       sasl password
-server         irc server (default: irc.libera.chat:6697)
-prefix         command prefix (default: !)
-nossl          disable ssl for irc
-nocheck	    disable ssl verification (e.g. self signed ssl)
-nominatim      nominatim server (default: http://nominatim.openstreetmap.org)
-nolimit        disable flood kick protection
-colors         enable irc colors
-bind           bind to host address
-debug          debug irc traffic
-yaml           yaml config file
```

### Specifying channel key

-channels "#channelname:channelkey, #channelname2:channelkey2"

### Yaml config file

See example config: [settings.yaml](settings.yaml)

Useful if you wanna run multiple bot instances across different IRC hosts
