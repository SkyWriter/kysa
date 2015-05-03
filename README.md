# Pocket IPTV for Raspberry Pi (and possibly others?)

## OMG, Why?

At some point I wondered if I could watch IPTV on my Raspberry Pi with
no hassle. Unsurprisingly, I've stumbled upon Kodi, which I immediately
installed and put to the test. After a few days of fiddling with it, I
figured that it was just too advanced for my use case.

All I wanted to do is feed the M3U URL, and IPTV just works (tm).

That's why I've built Kysa TV.

## Building

It's backend is built in Go and requires compilation. I personally
prefer not to wait eons to compile it on ARM, and do
cross-compliation with Gox. That's where ``build.sh`` script comes very
handy.

The script yields ``mytv_linux_arm`` file which is suitable to be
executed on your Raspberry Pi ARM.

## Installation

Just copy the whole directory (including the file compiled) to your Pi
and you're set.

## Running

Do the following on your Raspberry:

     $ sudo ./mytv_linux_arm http://hostname/playlist.m3u

Then use any browser to navigate to your new TV remote URL:
http://raspberry/.

There comes auto-populated from M3U list of channels to be switched. You
can also type something in the edit box to filter.

![](https://cloud.githubusercontent.com/assets/105392/7446208/0419eda6-f1f3-11e4-803b-825cba8018ed.png)
