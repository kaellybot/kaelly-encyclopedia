package constants

import amqp "github.com/kaellybot/kaelly-amqp"

const (
	DofusDudeGame              = "dofus2"
	DofusDudeDefaultLanguage   = "en"
	DofusDudeAlmanaxDateFormat = "2006-01-02"
	DofusDudeAlmanaxSizeLimit  = 35
	DofusDudeLimit             = 25
)

func GetSupportedTypeEnums() []string {
	return []string{
		"shield",
		"hat",
		"cloak",
		"amulet",
		"ring",
		"belt",
		"boots",
		"axe",
		"bow",
		"dagger",
		"hammer",
		"lance",
		"pickaxe",
		"scythe",
		"shovel",
		"staff",
		"sword",
		"wand",
		"dofus",
		"prysmaradite",
		"trophy",
		"pet",
		"petsmount",
		"mount",
	}
}

func GetLanguages() map[amqp.Language]string {
	return map[amqp.Language]string{
		amqp.Language_ANY: DofusDudeDefaultLanguage,
		amqp.Language_FR:  "fr",
		amqp.Language_EN:  "en",
		amqp.Language_ES:  "es",
		amqp.Language_DE:  "de",
	}
}

func GetDofusDudeSource() *amqp.Source {
	return &amqp.Source{
		Name: GetEncyclopediasSource().Name,
		Icon: GetEncyclopediasSource().Icon,
		Url:  GetEncyclopediasSource().URL,
	}
}
