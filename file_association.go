package main

type MarkdownAssociationStatus struct {
	Supported  bool `json:"supported"`
	Registered bool `json:"registered"`
	Default    bool `json:"default"`
}

func (a *App) GetMarkdownAssociationStatus() (MarkdownAssociationStatus, error) {
	return platformMarkdownAssociationStatus()
}

func (a *App) OpenMarkdownDefaultAppsSettings() error {
	return platformOpenMarkdownDefaultAppsSettings()
}
