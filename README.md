# Bot Detector

Automatically detect and kick bots & cheaters in TF2. 

## Warning

This is very early in development, expect bugs.

## What about [TF2 Bot Detector](https://github.com/PazerOP/tf2_bot_detector)?

If it works for you, feel free to keep using it, active development however has stopped. bd supports 
importing and exporting TF2BD player and rule lists to help ease adoption to this new tool. His tool is
quite difficult to hack on, so one of the goals of this project was to simplify that to encourage more
outside contributions.

## Current & Planned Features

- [x] Automatically download updated remote TF2BD lists
  - [x] Rules
  - [x] Players
- [ ] Cool logo
- [x] Custom 3rd party links
- [x] Discord rich presence
- [x] Fetch profile summary and ban info from steam web api
- [ ] Detection Methods
  - [x] Steam ID
  - [x] Name Pattern
  - [x] Avatar Pattern
  - [ ] Multi match
- [x] Translations
  - [x] English
  - [x] Russian
- [ ] WebGUI / Widget 
  - [x] Player status display list
  - [x] Current game chat dialogue 
    - [x] Send in=game chat messages
  - [ ] Player profile panel
    - [ ] Show the highest level of UGC/ETF2L/RGL league history achieved
    - [ ] Logs.tf count
  - [x] Player all-time chat history dialogue
  - [x] Player all-time name history dialogue
  - [x] Track all-time k:d against players
  - [x] External link configuration dialogue
  - [x] List configuration dialogue
  - [x] Settings dialogue
  - [ ] Rule creator & tester
  - [x] Auto start TF2 on launch & auto quit on game close.

## Installation

Check the [releases](https://github.com/leighmacdonald/bd/releases) page for latest binaries. There is currently
no installers so just extract anywhere and run. All data will be stored in the same location.

## Development

To build for linux, install the prerequisite libraries first.

- Linux (debian/ubuntu)
  - `sudo apt-get install gcc libgtk-3-dev libayatana-appindicator3-dev make`

Note that some distros may require the `libxapp-dev` package to be installed as well.

Checkout source

    # New checkout
    git clone --recurse-submodules -j8 git://github.com/leighmacdonald/bd.git && cd bd
    
    # (or) Existing repo and/or Old git version
    git clone git://github.com/leighmacdonald/bd.git
    cd bd && git submodule update --init --recursive

Linkers and static analysers 

    make check

Run tests

    make test

Or, Build it and run it.

    go build && ./bd

Releasing with cgo + windows is a bit annoying, so we just use wsl for now. Feel free to improve via pr.
    
    (wsl) $ goreleaser release --clean --split
    (win) $ goreleaser release --clean --split
    (wsl) $ cp -rv /mnt/c/projects/bd/dist/windows dist/
    (wsl) $ goreleaser continue --merge

### Editing Translations

#### New Language

1. Make an empty translation file. e.g. for french: `internal/tr/translate.fr.yaml`
2. Generate translation file: `goi18n merge .\internal\tr\active.en.yaml .\internal\tr\translate.fr.yaml`
3. Edit `.\internal\tr\translate.fr.yaml` with translations
4. Rename `.\internal\tr\translate.fr.yaml` to `.\internal\tr\active.fr.yaml`
5. Merge changes: `make tr_merge`

#### Updated Messages

1. `make tr_extract`
2. `make tr_gen_translate`
3. Edit updated `interlal/tr/translate.*.yaml` files
4. `make tr_merge`

See [go-i18n](https://github.com/nicksnyder/go-i18n) for more detailed instructions.
