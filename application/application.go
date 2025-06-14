package application

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	almanaxRepo "github.com/kaellybot/kaelly-encyclopedia/repositories/almanaxes"
	equipmentRepo "github.com/kaellybot/kaelly-encyclopedia/repositories/equipments"
	"github.com/kaellybot/kaelly-encyclopedia/repositories/games"
	setRepo "github.com/kaellybot/kaelly-encyclopedia/repositories/sets"
	"github.com/kaellybot/kaelly-encyclopedia/repositories/weapons"
	"github.com/kaellybot/kaelly-encyclopedia/services/almanaxes"
	"github.com/kaellybot/kaelly-encyclopedia/services/encyclopedias"
	"github.com/kaellybot/kaelly-encyclopedia/services/equipments"
	"github.com/kaellybot/kaelly-encyclopedia/services/news"
	"github.com/kaellybot/kaelly-encyclopedia/services/sets"
	"github.com/kaellybot/kaelly-encyclopedia/services/sources"
	"github.com/kaellybot/kaelly-encyclopedia/services/stores"
	"github.com/kaellybot/kaelly-encyclopedia/utils/databases"
	"github.com/kaellybot/kaelly-encyclopedia/utils/insights"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New() (*Impl, error) {
	// misc
	broker := amqp.New(constants.RabbitMQClientID, viper.GetString(constants.RabbitMQAddress),
		amqp.WithBindings(encyclopedias.GetBinding()))
	db := databases.New()
	if errDB := db.Run(); errDB != nil {
		return nil, errDB
	}

	probes := insights.NewProbes(broker.IsConnected, db.IsConnected)
	prom := insights.NewPrometheusMetrics()

	// Create scheduler with Europe/Paris timezone
	frenchLocation, err := time.LoadLocation(constants.FrenchTimezone)
	if err != nil {
		return nil, err
	}

	// Since we have winter/summer hours, UTC location cannot be used easily.
	scheduler, errScheduler := gocron.NewScheduler(gocron.WithLocation(frenchLocation))
	if errScheduler != nil {
		return nil, errScheduler
	}

	// Repositories
	almanaxRepo := almanaxRepo.New(db)
	equipmentRepo := equipmentRepo.New(db)
	weaponRepo := weapons.New(db)
	setRepo := setRepo.New(db)
	gameRepo := games.New(db)

	// services
	storeService := stores.New()
	equipmentService, errEquipment := equipments.New(equipmentRepo, weaponRepo)
	if errEquipment != nil {
		return nil, errEquipment
	}

	sourceService, errSource := sources.New(scheduler, storeService, gameRepo)
	if errSource != nil {
		return nil, errSource
	}

	newsService := news.New(broker, sourceService)
	almanaxService, errAlmanax := almanaxes.New(scheduler, frenchLocation,
		almanaxRepo, sourceService, newsService)
	if errAlmanax != nil {
		return nil, errAlmanax
	}

	setService, errSet := sets.New(setRepo, newsService, sourceService, equipmentService)
	if errSet != nil {
		return nil, errSet
	}

	encyclopediaService := encyclopedias.New(broker, sourceService,
		almanaxService, equipmentService, setService)

	return &Impl{
		broker:              broker,
		scheduler:           scheduler,
		db:                  db,
		probes:              probes,
		prom:                prom,
		encyclopediaService: encyclopediaService,
	}, nil
}

func (app *Impl) Run() error {
	app.probes.ListenAndServe()
	app.prom.ListenAndServe()

	errBroker := app.broker.Run()
	if errBroker != nil {
		return errBroker
	}

	app.scheduler.Start()
	for _, job := range app.scheduler.Jobs() {
		scheduledTime, err := job.NextRun()
		if err == nil {
			log.Info().Msgf("%v scheduled at %v", job.Name(), scheduledTime)
		}
	}

	return app.encyclopediaService.Consume()
}

func (app *Impl) Shutdown() {
	if err := app.scheduler.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Cannot shutdown scheduler, continuing...")
	}

	app.broker.Shutdown()
	app.db.Shutdown()
	app.prom.Shutdown()
	app.probes.Shutdown()
	log.Info().Msgf("Application is no longer running")
}
