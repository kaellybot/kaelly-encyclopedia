package constants

type Source struct {
	Name string
	Icon string
	URL  string
}

func GetEncyclopediasSource() Source {
	return Source{
		Name: "dofusdude",
		Icon: "https://avatars.githubusercontent.com/u/82651571",
		URL:  "https://github.com/dofusdude",
	}
}
