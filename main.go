package main

import (
	"context"
	"encoding/json"
	"github.com/golang-migrate/migrate/v4"
	"github.com/leighmacdonald/bd/internal/cache"
	"github.com/leighmacdonald/bd/internal/detector"
	"github.com/leighmacdonald/bd/internal/model"
	"github.com/leighmacdonald/bd/internal/store"
	"github.com/leighmacdonald/bd/internal/ui"
	"github.com/leighmacdonald/bd/pkg/rules"
	"github.com/leighmacdonald/bd/pkg/util"
	"github.com/leighmacdonald/steamweb"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"log"
	"os"
)

var (
	// Build info
	version string = "master"
	commit  string = "latest"
	date    string = "n/a"
	builtBy string = "src"
)

func main() {
	ctx := context.Background()
	settings, errSettings := model.NewSettings()
	if errSettings != nil {
		log.Panicf("Failed to initialize settings: %v", errSettings)
	}

	localRules := rules.NewRuleSchema()
	localPlayersList := rules.NewPlayerListSchema()
	if errReadSettings := settings.ReadDefaultOrCreate(); errReadSettings != nil {
		log.Println(errReadSettings)
	}
	// Try and load our existing custom players/rules
	if util.Exists(settings.LocalPlayerListPath()) {
		input, errInput := os.Open(settings.LocalPlayerListPath())
		if errInput != nil {
			log.Printf("Failed to open local player list\n")
		} else {
			if errRead := json.NewDecoder(input).Decode(&localPlayersList); errRead != nil {
				log.Printf("Failed to parse local player list: %v\n", errRead)
			}
			util.LogClose(input)
		}
	}
	if util.Exists(settings.LocalRulesListPath()) {
		input, errInput := os.Open(settings.LocalRulesListPath())
		if errInput != nil {
			log.Printf("Failed to open local rules list\n")
		} else {
			if errRead := json.NewDecoder(input).Decode(&localRules); errRead != nil {
				log.Printf("Failed to parse local rules list: %v\n", errRead)
			}
			util.LogClose(input)
		}
	}
	engine, ruleEngineErr := rules.New(&localRules, &localPlayersList)
	if ruleEngineErr != nil {
		log.Panicf("Failed to setup rules engine: %v\n", ruleEngineErr)
	}
	if settings.GetAPIKey() != "" {
		if errAPIKey := steamweb.SetKey(settings.GetAPIKey()); errAPIKey != nil {
			log.Printf("Failed to set steam api key: %v\n", errAPIKey)
		}
	}
	dataStore := store.New(settings.DBPath())
	if errMigrate := dataStore.Init(); errMigrate != nil && !errors.Is(errMigrate, migrate.ErrNoChange) {
		log.Printf("Failed to migrate database: %v\n", errMigrate)
		os.Exit(1)
	}
	defer util.LogClose(dataStore)
	bd := detector.New(settings, dataStore, engine, cache.New(settings.ConfigRoot(), model.DurationCacheTimeout))
	gui := ui.New(ctx, &bd, settings, dataStore, model.Version{
		Version: version,
		Commit:  commit,
		Date:    date,
		BuiltBy: builtBy,
	})
	gui.Start(ctx)
}
