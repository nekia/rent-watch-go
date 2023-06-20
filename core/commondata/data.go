package commondata

type CrawlReq struct {
	SiteName string `json:"siteName,omitempty"`
}

type CrawlResp struct {
	Url string `json:"url,omitempty"`
}

type FloorLevel struct {
	FloorLevel    int `json:"floorLevel,omitempty"`
	FloorTopLevel int `json:"floorTopLevel,omitempty"`
}
type ScanResp struct {
	Address    string     `json:"address,omitempty"`
	Price      int        `json:"price,omitempty"`
	Size       float64    `json:"size,omitempty"`
	FloorLevel FloorLevel `json:"floorLevel,omitempty"`
	Location   string     `json:"location,omitempty"`
	BuiltYear  int        `json:"builtYear,omitempty"`
	IsPetOK    bool       `json:"isPetOK,omitempty"`
}

type ScanReq struct {
	SiteName string `json:"siteName,omitempty"`
	Url      string `json:"url,omitempty"`
}
