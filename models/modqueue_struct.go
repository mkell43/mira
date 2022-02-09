package models

type ModQueueListing struct {
	Kind string              `json:"kind"`
	Data ModQueueListingData `json:"data"`
}

type ModQueueListingData struct {
	Modhash  string                 `json:"modhash"`
	Dist     float64                `json:"dist"`
	Children []ModQueueListingChild `json:"children"`
}

type ModQueueListingChild struct {
	Kind string               `json:"kind"`
	Data PostListingChildData `json:"data"`
}
