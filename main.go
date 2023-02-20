package main

import (
	"context"
	"github.com/golang-migrate/migrate/v4"
	"github.com/leighmacdonald/bd/model"
	_ "github.com/leighmacdonald/bd/translations"
	"github.com/leighmacdonald/bd/ui"
	"github.com/leighmacdonald/steamweb"
	"github.com/pkg/errors"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	settings := model.NewSettings()
	localRules := newRuleSchema()
	localPlayersList := newPlayerListSchema()

	if errReadSettings := settings.ReadDefaultOrCreate(); errReadSettings != nil {
		log.Println(errReadSettings)
	}

	// Try and load our existing custom players/rules
	if exists(settings.LocalPlayerListPath()) {
		input, errInput := os.Open(settings.LocalPlayerListPath())
		if errInput != nil {
			log.Printf("Failed to open local player list\n")
		} else {
			if errRead := parsePlayerSchema(input, &localPlayersList); errRead != nil {
				log.Printf("Failed to parse local player list: %v\n", errRead)
			}
			logClose(input)
		}
	}
	if exists(settings.LocalRulesListPath()) {
		input, errInput := os.Open(settings.LocalRulesListPath())
		if errInput != nil {
			log.Printf("Failed to open local rules list\n")
		} else {
			if errRead := parseRulesList(input, &localRules); errRead != nil {
				log.Printf("Failed to parse local rules list: %v\n", errRead)
			}
			logClose(input)
		}
	}
	rulesEngine, ruleEngineErr := newRulesEngine(&localRules, &localPlayersList)
	if ruleEngineErr != nil {
		log.Panicf("Failed to setup rules engine: %v\n", ruleEngineErr)
	}

	if settings.ApiKey != "" {
		if errApiKey := steamweb.SetKey(settings.ApiKey); errApiKey != nil {
			log.Printf("Failed to set steam api key: %v\n", errApiKey)
		}
	}
	store := newSqliteStore(settings.DBPath())
	if errMigrate := store.Init(); errMigrate != nil && !errors.Is(errMigrate, migrate.ErrNoChange) {
		log.Printf("Failed to migrate database: %v\n", errMigrate)
		os.Exit(1)
	}
	defer store.Close()
	bd := New(ctx, &settings, store, rulesEngine)

	gui := ui.New(ctx, &settings)
	defer bd.Shutdown()
	bd.AttachGui(gui)
	go bd.start()
	gui.Start()
}
