package sets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/dofusdude/dodugo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	repository "github.com/kaellybot/kaelly-encyclopedia/repositories/sets"
	"github.com/kaellybot/kaelly-encyclopedia/services/equipments"
	"github.com/kaellybot/kaelly-encyclopedia/services/news"
	"github.com/kaellybot/kaelly-encyclopedia/services/sources"
	"github.com/rs/zerolog/log"
)

func New(repository repository.Repository, newsService news.Service,
	sourceService sources.Service, equipmentService equipments.Service) (*Impl, error) {
	service := Impl{
		newsService:      newsService,
		sourceService:    sourceService,
		equipmentService: equipmentService,
		sets:             make(map[int64]entities.Set),
		repository:       repository,
	}

	errDB := service.loadSetsFromDB()
	if errDB != nil {
		return nil, errDB
	}

	service.sourceService.ListenGameEvent(service.syncSetImages)

	return &service, nil
}

func (service *Impl) GetSetByDofusDude(id int64) (entities.Set, bool) {
	service.mu.RLock()
	defer service.mu.RUnlock()
	item, found := service.sets[id]
	return item, found
}

func (service *Impl) loadSetsFromDB() error {
	service.mu.Lock()
	defer service.mu.Unlock()

	sets, err := service.repository.GetSets()
	if err != nil {
		return err
	}

	log.Info().
		Int(constants.LogEntityCount, len(sets)).
		Msgf("Sets loaded")

	service.sets = make(map[int64]entities.Set)
	for _, set := range sets {
		service.sets[int64(set.ID)] = set
	}

	return nil
}

func (service *Impl) syncSetImages(_ string) {
	log.Info().Msgf("Syncing set images...")

	// Checking sets differences
	createdSets, updatedSetIDs, deletedSetIDs, errCheck := service.checkSetImages()
	if errCheck != nil {
		log.Error().Err(errCheck).Msgf("Cannot sync sets images, trying later...")
		return
	}

	if len(createdSets) == 0 && len(updatedSetIDs) == 0 && len(deletedSetIDs) == 0 {
		log.Info().Msgf("Set images are all up-to-date")
		return
	}

	log.Info().Msgf("Some set images are not up-to-date, publishing dedicated news and updating database in consequence")
	createdSetIDs := make([]string, len(createdSets))
	for i, set := range createdSets {
		createdSetIDs[i] = fmt.Sprintf("%v", set.ID)
	}

	// At this point, we've seen differences. Publishing news...
	service.newsService.PublishSetNews(createdSetIDs, updatedSetIDs, deletedSetIDs)

	// Saving into DB
	errSave := service.repository.Sync(createdSets, updatedSetIDs, deletedSetIDs)
	if errSave != nil {
		log.Error().Err(errSave).Msgf("Cannot save sets images, trying later...")
		return
	}

	// And reload set images
	errLoad := service.loadSetsFromDB()
	if errLoad != nil {
		log.Error().Err(errSave).Msgf("Cannot reload sets images, continuing...")
		return
	}
}

func (service *Impl) checkSetImages() ([]entities.Set, []string, []string, error) {
	ctx := context.Background()

	checkedSets := make(map[int64]bool)
	for setID := range service.sets {
		checkedSets[setID] = false
	}

	sourceSets, errGet := service.sourceService.GetSets(ctx)
	if errGet != nil {
		return nil, nil, nil, errGet
	}

	createdSets := make([]entities.Set, 0)
	updatedSetIDs := make([]string, 0)
	deletedSetIDs := make([]string, 0)
	// Checking if existing sets have changed or just have been created
	for _, sourceSet := range sourceSets {
		setID := int64(sourceSet.GetAnkamaId())
		setIDStr := fmt.Sprintf("%v", setID)
		sourceSetHash := getSetHash(sourceSet)
		if set, found := service.sets[setID]; found {
			if sourceSetHash != set.Hash {
				updatedSetIDs = append(updatedSetIDs, setIDStr)
			}

			checkedSets[setID] = true
		} else {
			createdSets = append(createdSets, entities.Set{
				ID:        sourceSet.GetAnkamaId(),
				Game:      amqp.Game_DOFUS_GAME,
				Hash:      sourceSetHash,
				Icon:      fmt.Sprintf(constants.SetImageBaseURL, setID),
				IsCurrent: false,
			})
		}
	}

	// Checking potential deleted sets
	for setID, checked := range checkedSets {
		if !checked {
			deletedSetIDs = append(deletedSetIDs, fmt.Sprintf("%v", setID))
		}
	}

	return createdSets, updatedSetIDs, deletedSetIDs, nil
}

func getSetHash(set dodugo.ListEquipmentSet) string {
	slices.Sort(set.GetEquipmentIds())
	itemIDs := make([]string, len(set.GetEquipmentIds()))
	for i, itemID := range set.GetEquipmentIds() {
		itemIDs[i] = fmt.Sprint(itemID)
	}

	rawHash := strings.Join(itemIDs, "-")
	hash := sha256.Sum256([]byte(rawHash))
	return hex.EncodeToString(hash[:])
}
