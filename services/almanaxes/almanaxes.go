package almanaxes

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/go-co-op/gocron/v2"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	"github.com/kaellybot/kaelly-encyclopedia/models/mappers"
	repository "github.com/kaellybot/kaelly-encyclopedia/repositories/almanaxes"
	"github.com/kaellybot/kaelly-encyclopedia/services/news"
	"github.com/kaellybot/kaelly-encyclopedia/services/sources"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New(scheduler gocron.Scheduler, frenchLocation *time.Location, repository repository.Repository,
	sourceService sources.Service, newsService news.Service) (*Impl, error) {
	service := Impl{
		frenchLocation: frenchLocation,
		almanaxes:      make(map[string][]entities.Almanax),
		sourceService:  sourceService,
		newsService:    newsService,
		repository:     repository,
	}

	errDB := service.loadAlmanaxEffectsFromDB()
	if errDB != nil {
		return nil, errDB
	}

	service.sourceService.ListenGameEvent(service.reconcileDofusDudeIDs)

	_, errJob := scheduler.NewJob(
		gocron.CronJob(viper.GetString(constants.AlmanaxCronTab), true),
		gocron.NewTask(func() { service.dispatchDailyAlmanax() }),
		gocron.WithName("Dispatch daily almanax"),
	)
	if errJob != nil {
		return nil, errJob
	}

	return &service, nil
}

func (service *Impl) GetLocation() *time.Location {
	return service.frenchLocation
}

func (service *Impl) GetDatesByAlmanaxEffect(dofusDudeEffectID string) []time.Time {
	now := time.Now().UTC()
	dates := make([]time.Time, 0)
	entities, found := service.almanaxes[dofusDudeEffectID]
	if !found {
		return dates
	}

	for _, entity := range entities {
		year := now.Year()
		if time.Month(entity.Month) < now.Month() ||
			time.Month(entity.Month) == now.Month() && entity.Day < now.Day() {
			year++
		}

		date := time.Date(year, time.Month(entity.Month), entity.Day, 0, 0, 0, 0, time.UTC)

		// Check if created date matches (29th february is not available every year for ex.)
		if date.Year() != year || int(date.Month()) != entity.Month || date.Day() != entity.Day {
			continue
		}

		dates = append(dates, date)
	}

	// Sorted by ASC
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	return dates
}

func (service *Impl) loadAlmanaxEffectsFromDB() error {
	almanaxes, err := service.repository.GetAlmanaxes()
	if err != nil {
		return err
	}

	log.Info().
		Int(constants.LogEntityCount, len(almanaxes)).
		Msgf("Almanaxes loaded")

	for _, almanax := range almanaxes {
		effects := service.almanaxes[almanax.DofusDudeEffectID]
		service.almanaxes[almanax.DofusDudeEffectID] = append(effects, almanax)
	}

	return nil
}

func (service *Impl) dispatchDailyAlmanax() {
	log.Info().Msgf("Dispatching daily almanax...")
	almanaxes := make([]*amqp.NewsAlmanaxMessage_I18NAlmanax, 0)
	for _, value := range amqp.Language_value {
		lg := amqp.Language(value)

		// ignore default value like ANY, NONE, etc.
		if lg == amqp.Language_ANY {
			continue
		}

		day := time.Now().In(service.frenchLocation)
		dofusDudeLg, found := constants.GetLanguages()[lg]
		if !found {
			log.Warn().Msgf("Cannot retrieve DofusDude language from amqp.Locale '%v',"+
				" continuing without this almanax", lg)
			continue
		}
		almanax, err := service.sourceService.GetAlmanaxByDate(context.Background(), day, dofusDudeLg)
		if err != nil {
			log.Warn().Err(err).
				Msgf("Cannot retrieve almanax from DofusDude (lg=%v), continuing without it", lg)
			continue
		}

		almanaxes = append(almanaxes, &amqp.NewsAlmanaxMessage_I18NAlmanax{
			Almanax: mappers.MapAlmanax(almanax, service.sourceService),
			Locale:  lg,
		})
	}

	service.newsService.PublishAlmanaxNews(almanaxes)
}

func (service *Impl) reconcileDofusDudeIDs(_ string) {
	log.Info().Msgf("Reconciling almanax DofusDude IDs...")
	ctx := context.Background()
	year := time.Now().Year()

	almanaxEntities, errDB := service.repository.GetAlmanaxes()
	if errDB != nil {
		log.Error().Err(errDB).Msgf("Cannot retrieve almanaxes from database, trying later...")
		return
	}

	var updatedCount int
	var errorCount int
	for _, almanaxEntity := range almanaxEntities {
		updated, errRec := service.reconcileDofusDudeID(ctx, almanaxEntity, year)
		if errRec != nil {
			log.Warn().Err(errRec).
				Str(constants.LogDate, fmt.
					Sprintf("%v-%v-%v", year, almanaxEntity.Month, almanaxEntity.Day)).
				Msgf("Error while reconciliating almanax, continuing without this date")
			errorCount++
			continue
		}

		if updated {
			updatedCount++
		}
	}

	if updatedCount == 0 {
		log.Info().
			Int(constants.LogEntityCount, len(almanaxEntities)).
			Msgf("Almanax days are all up-to-date")
		return
	}

	log.Info().
		Int(constants.LogEntityCount, updatedCount).
		Msg("Almanax dates reconciliated!")

	errLoad := service.loadAlmanaxEffectsFromDB()
	log.Warn().Err(errLoad).Msg("Could not reload almanax from DB, please restart to take them in account")
}

func (service *Impl) reconcileDofusDudeID(ctx context.Context, entity entities.Almanax, year int,
) (bool, error) {
	day := time.Date(year, time.Month(entity.Month), entity.Day, 0, 0, 0, 0, time.UTC)
	dodugoAlmanax, errGet := service.sourceService.
		GetAlmanaxByDate(ctx, day, constants.DofusDudeDefaultLanguage)
	if errGet != nil {
		return false, errGet
	}

	if dodugoAlmanax == nil {
		return false, errNotFound
	}

	if dodugoAlmanax.Bonus.Type.GetId() != entity.DofusDudeEffectID {
		entity.DofusDudeEffectID = dodugoAlmanax.Bonus.Type.GetId()
		return true, service.repository.Save(entity)
	}

	return false, nil
}
