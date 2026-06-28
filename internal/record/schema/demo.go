package schema

type DemoFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type Demo struct {
	Name      string     `json:"name"`
	CreatedAt int64      `json:"created_at"`
	Files     []DemoFile `json:"files"`
}
