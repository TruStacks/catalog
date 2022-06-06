package main

type concourse struct {
	baseComponent
}

func init() {
	catalog.addComponent("concourse", &concourse{
		baseComponent{
			Repo:    "https://concourse-charts.storage.googleapis.com",
			Chart:   "concourse",
			Version: "16.1.22",
			Hooks:   make([]string, 0),
		},
	})
}
