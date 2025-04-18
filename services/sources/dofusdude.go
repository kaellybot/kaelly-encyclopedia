package sources

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dofusdude/dodugo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/kaellybot/kaelly-encyclopedia/utils/conversions"
	"github.com/rs/zerolog/log"
)

func (service *Impl) GetItemType(itemType string) amqp.ItemType {
	amqpItemType, found := service.itemTypes[itemType]
	if !found {
		log.Warn().
			Str(constants.LogItemType, itemType).
			Msgf("Cannot find dofusDude itemType match, returning amqp.ItemType_ANY_ITEM")
		return amqp.ItemType_ANY_ITEM_TYPE
	}

	return amqpItemType
}

func (service *Impl) SearchAnyItems(ctx context.Context, query,
	language string) ([]dodugo.GameSearch, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	var items []dodugo.GameSearch
	key := buildListKey(item, query, language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &items) {
		resp, r, err := service.dofusDudeClient.
			GameAPI.
			GetGameSearch(ctx, language, constants.DofusDudeGame).
			Query(query).
			FilterSearchIndex(constants.GetSupportedSearchIndex()).
			FilterTypeNameId(constants.GetSupportedTypeEnums()).
			Limit(constants.DofusDudeLimit).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		items = resp
	}

	return items, nil
}

//nolint:dupl // Complicated to be more DRY.
func (service *Impl) GetConsumableByID(ctx context.Context, itemID int64, language string,
) (*dodugo.Resource, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	int32ItemID, errConv := conversions.Int64ToInt32(itemID)
	if errConv != nil {
		return nil, errConv
	}

	var dodugoItem *dodugo.Resource
	key := buildItemKey(item, fmt.Sprintf("%v", itemID), language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &dodugoItem) {
		resp, r, err := service.dofusDudeClient.ConsumablesAPI.
			GetItemsConsumablesSingle(ctx, language, int32ItemID, constants.DofusDudeGame).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		dodugoItem = resp
	}

	return dodugoItem, nil
}

func (service *Impl) SearchCosmetics(ctx context.Context, query,
	language string) ([]dodugo.ListItem, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	var items []dodugo.ListItem
	key := buildListKey(item, query, language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &items) {
		resp, r, err := service.dofusDudeClient.CosmeticsAPI.
			GetCosmeticsSearch(ctx, language, constants.DofusDudeGame).
			Query(query).Limit(constants.DofusDudeLimit).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		items = resp
	}

	return items, nil
}

func (service *Impl) GetCosmeticByQuery(ctx context.Context, query, language string,
) (*dodugo.Weapon, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	values, err := service.SearchCosmetics(ctx, query, language)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrNotFound
	}

	// We trust the omnisearch by taking the first one in the list
	resp, err := service.GetCosmeticByID(ctx, int64(values[0].GetAnkamaId()), language)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (service *Impl) GetCosmeticByID(ctx context.Context, itemID int64, language string,
) (*dodugo.Weapon, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	int32ItemID, errConv := conversions.Int64ToInt32(itemID)
	if errConv != nil {
		return nil, errConv
	}

	var dodugoItem *dodugo.Weapon
	key := buildItemKey(item, fmt.Sprintf("%v", itemID), language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &dodugoItem) {
		resp, r, err := service.dofusDudeClient.CosmeticsAPI.
			GetCosmeticsSingle(ctx, language, int32ItemID, constants.DofusDudeGame).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		isWeapon := false
		dodugoItem = &dodugo.Weapon{
			AnkamaId:               resp.AnkamaId,
			Name:                   resp.Name,
			Description:            resp.Description,
			Type:                   resp.Type,
			IsWeapon:               &isWeapon,
			Level:                  resp.Level,
			Pods:                   resp.Pods,
			ImageUrls:              resp.ImageUrls,
			Effects:                resp.Effects,
			Conditions:             resp.Conditions,
			Recipe:                 resp.Recipe,
			ParentSet:              resp.ParentSet,
			CriticalHitProbability: nil,
			CriticalHitBonus:       nil,
			MaxCastPerTurn:         nil,
			ApCost:                 nil,
			Range:                  nil,
		}
	}

	return dodugoItem, nil
}

func (service *Impl) SearchEquipments(ctx context.Context, query,
	language string) ([]dodugo.ListItem, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	var items []dodugo.ListItem
	key := buildListKey(item, query, language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &items) {
		resp, r, err := service.dofusDudeClient.EquipmentAPI.
			GetItemsEquipmentSearch(ctx, language, constants.DofusDudeGame).
			Query(query).Limit(constants.DofusDudeLimit).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		items = resp
	}

	return items, nil
}

func (service *Impl) GetEquipmentByQuery(ctx context.Context, query, language string,
) (*dodugo.Weapon, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	values, err := service.SearchEquipments(ctx, query, language)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrNotFound
	}

	// We trust the omnisearch by taking the first one in the list
	resp, err := service.GetEquipmentByID(ctx, int64(values[0].GetAnkamaId()), language)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

//nolint:dupl // Complicated to be more DRY.
func (service *Impl) GetEquipmentByID(ctx context.Context, itemID int64, language string,
) (*dodugo.Weapon, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	int32ItemID, errConv := conversions.Int64ToInt32(itemID)
	if errConv != nil {
		return nil, errConv
	}

	var dodugoItem *dodugo.Weapon
	key := buildItemKey(item, fmt.Sprintf("%v", itemID), language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &dodugoItem) {
		resp, r, err := service.dofusDudeClient.EquipmentAPI.
			GetItemsEquipmentSingle(ctx, language, int32ItemID, constants.DofusDudeGame).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		dodugoItem = resp
	}

	return dodugoItem, nil
}

//nolint:dupl // Complicated to be more DRY.
func (service *Impl) GetQuestItemByID(ctx context.Context, itemID int64, language string,
) (*dodugo.Resource, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	int32ItemID, errConv := conversions.Int64ToInt32(itemID)
	if errConv != nil {
		return nil, errConv
	}

	var dodugoItem *dodugo.Resource
	key := buildItemKey(item, fmt.Sprintf("%v", itemID), language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &dodugoItem) {
		resp, r, err := service.dofusDudeClient.QuestItemsAPI.
			GetItemQuestSingle(ctx, language, int32ItemID, constants.DofusDudeGame).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		dodugoItem = resp
	}

	return dodugoItem, nil
}

//nolint:dupl // Complicated to be more DRY.
func (service *Impl) GetResourceByID(ctx context.Context, itemID int64, language string,
) (*dodugo.Resource, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	int32ItemID, errConv := conversions.Int64ToInt32(itemID)
	if errConv != nil {
		return nil, errConv
	}

	var dodugoItem *dodugo.Resource
	key := buildItemKey(item, fmt.Sprintf("%v", itemID), language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &dodugoItem) {
		resp, r, err := service.dofusDudeClient.ResourcesAPI.
			GetItemsResourcesSingle(ctx, language, int32ItemID, constants.DofusDudeGame).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		dodugoItem = resp
	}

	return dodugoItem, nil
}

func (service *Impl) SearchMounts(ctx context.Context, query,
	language string) ([]dodugo.Mount, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	var items []dodugo.Mount
	key := buildListKey(item, query, language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &items) {
		resp, r, err := service.dofusDudeClient.MountsAPI.
			GetMountsSearch(ctx, language, constants.DofusDudeGame).
			Query(query).Limit(constants.DofusDudeLimit).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		items = resp
	}

	return items, nil
}

func (service *Impl) GetMountByQuery(ctx context.Context, query, language string,
) (*dodugo.Mount, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	values, err := service.SearchMounts(ctx, query, language)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrNotFound
	}

	// We trust the omnisearch by taking the first one in the list
	resp, err := service.GetMountByID(ctx, int64(values[0].GetAnkamaId()), language)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

//nolint:dupl // Complicated to be more DRY.
func (service *Impl) GetMountByID(ctx context.Context, itemID int64, language string,
) (*dodugo.Mount, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	int32ItemID, errConv := conversions.Int64ToInt32(itemID)
	if errConv != nil {
		return nil, errConv
	}

	var dodugoItem *dodugo.Mount
	key := buildItemKey(item, fmt.Sprintf("%v", itemID), language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &dodugoItem) {
		resp, r, err := service.dofusDudeClient.MountsAPI.
			GetMountsSingle(ctx, language, int32ItemID, constants.DofusDudeGame).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		dodugoItem = resp
	}

	return dodugoItem, nil
}

func (service *Impl) SearchSets(ctx context.Context, query,
	language string) ([]dodugo.ListEquipmentSet, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	var sets []dodugo.ListEquipmentSet
	key := buildListKey(set, query, language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &sets) {
		resp, r, err := service.dofusDudeClient.SetsAPI.
			GetSetsSearch(ctx, language, constants.DofusDudeGame).
			Query(query).Limit(constants.DofusDudeLimit).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		sets = resp
	}

	return sets, nil
}

func (service *Impl) GetSetByQuery(ctx context.Context, query, language string,
) (*dodugo.EquipmentSet, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	values, err := service.SearchSets(ctx, query, language)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrNotFound
	}

	// We trust the omnisearch by taking the first one in the list
	resp, err := service.GetSetByID(ctx, int64(values[0].GetAnkamaId()), language)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

//nolint:dupl // Complicated to be more DRY.
func (service *Impl) GetSetByID(ctx context.Context, setID int64, language string,
) (*dodugo.EquipmentSet, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	int32ItemID, errConv := conversions.Int64ToInt32(setID)
	if errConv != nil {
		return nil, errConv
	}

	var dodugoSet *dodugo.EquipmentSet
	key := buildItemKey(set, fmt.Sprintf("%v", setID), language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &dodugoSet) {
		resp, r, err := service.dofusDudeClient.SetsAPI.
			GetSetsSingle(ctx, language, int32ItemID, constants.DofusDudeGame).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		dodugoSet = resp
	}

	return dodugoSet, nil
}

// Returns sets with minimal informations. No cache applied here.
func (service *Impl) GetSets(ctx context.Context) ([]dodugo.ListEquipmentSet, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	resp, r, err := service.dofusDudeClient.SetsAPI.
		GetSetsList(ctx, constants.DofusDudeDefaultLanguage, constants.DofusDudeGame).
		PageNumber(1).PageSize(-1).FieldsSet([]string{"equipment_ids"}).
		Execute()
	if err != nil && r == nil {
		return nil, err
	}
	defer r.Body.Close()

	return resp.GetSets(), nil
}

func (service *Impl) SearchAlmanaxEffects(ctx context.Context, query,
	language string) ([]dodugo.GetMetaAlmanaxBonuses200ResponseInner, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	var effects []dodugo.GetMetaAlmanaxBonuses200ResponseInner
	key := buildListKey(almanaxEffect, query, language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &effects) {
		resp, r, err := service.dofusDudeClient.MetaAPI.
			GetMetaAlmanaxBonusesSearch(ctx, language).
			Query(query).
			Limit(constants.DofusDudeLimit).
			Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		effects = resp
	}

	return effects, nil
}

func (service *Impl) GetAlmanaxByDate(ctx context.Context, date time.Time, language string,
) (*dodugo.Almanax, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	var dodugoAlmanax *dodugo.Almanax
	dodugoAlmanaxDate := date.Format(constants.DofusDudeAlmanaxDateFormat)
	key := buildItemKey(almanax, dodugoAlmanaxDate, language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &dodugoAlmanax) {
		resp, r, err := service.dofusDudeClient.AlmanaxAPI.
			GetAlmanaxDate(ctx, language, dodugoAlmanaxDate).Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		dodugoAlmanax = resp
	}

	if dodugoAlmanax == nil {
		currentYear := time.Now().Year()
		if currentYear != date.Year() {
			log.Warn().
				Str(constants.LogDate, dodugoAlmanaxDate).
				Msgf("DofusDude API returns 404 NOT_FOUND for specific date, continuing with closest date...")
			fallbackDate := time.Date(currentYear, date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
			fallbackAlmanax, errFallback := service.GetAlmanaxByDate(ctx, fallbackDate, language)
			if fallbackAlmanax != nil {
				fallbackAlmanax.SetDate(dodugoAlmanaxDate)
				service.putElementToCache(ctx, key, fallbackAlmanax)
			}
			return fallbackAlmanax, errFallback
		}

		log.Error().
			Str(constants.LogDate, dodugoAlmanaxDate).
			Msgf("DofusDude API returns 404 NOT_FOUND for a close date, continuing with nil almanax...")
	}

	return dodugoAlmanax, nil
}

func (service *Impl) GetAlmanaxByRange(ctx context.Context, daysDuration int64, language string,
) ([]dodugo.Almanax, error) {
	ctx, cancel := context.WithTimeout(ctx, service.httpTimeout)
	defer cancel()

	int32DaysDuration, errConv := conversions.Int64ToInt32(daysDuration)
	if errConv != nil {
		return nil, errConv
	}

	var dodugoAlmanax []dodugo.Almanax
	dodugoAlmanaxDate := time.Now().Format(constants.DofusDudeAlmanaxDateFormat)
	key := buildItemKey(almanaxRange, fmt.Sprintf("%v_%v", dodugoAlmanaxDate, daysDuration),
		language, constants.GetEncyclopediasSource().Name)
	if !service.getElementFromCache(ctx, key, &dodugoAlmanax) {
		resp, r, err := service.dofusDudeClient.AlmanaxAPI.
			GetAlmanaxRange(ctx, language).
			RangeSize(int32DaysDuration).
			Execute()
		if err != nil && (r == nil || r.StatusCode != http.StatusNotFound) {
			return nil, err
		}
		defer r.Body.Close()
		service.putElementToCache(ctx, key, resp)
		dodugoAlmanax = resp
	}

	return dodugoAlmanax, nil
}
