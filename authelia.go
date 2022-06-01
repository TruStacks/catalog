package main

type authelia struct {
	baseComponent
}

func init() {
	catalog.addComponent("authelia", &authelia{
		baseComponent{
			Repo:    "https://charts.authelia.com",
			Chart:   "authelia/authelia",
			Version: "",
		},
	})
}
