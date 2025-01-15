package llama

const (
	FullPageScreenshot = "full_page_screenshot"
	Heading            = "heading"
	Table              = "table"
	Text               = "text"
)

type LlamaParse struct {
	Pages []LlamaPage `json:"pages"`
}

type LlamaPage struct {
	Page   int          `json:"page"`
	Text   string       `json:"text"`
	Md     string       `json:"md"`
	Images []LlamaImage `json:"images"`
	Items  []LlamaItem  `json:"items"`
	Status string       `json:"status"`
}

type LlamaImage struct {
	Name           string  `json:"name"`
	Height         float64 `json:"height"`
	Width          float64 `json:"width"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
	OriginalWidth  float64 `json:"original_width"`
	OriginalHeight float64 `json:"original_height"`
	Type           string  `json:"type"`
}

type LlamaItem struct {
	Type           string     `json:"type"`
	Lvl            int        `json:"lvl,omitempty"`
	Value          string     `json:"value,omitempty"`
	Md             string     `json:"md"`
	BBox           BBox       `json:"bBox,omitempty"`
	Rows           [][]string `json:"rows,omitempty"`
	IsPerfectTable bool       `json:"isPerfectTable,omitempty"`
	Csv            string     `json:"csv,omitempty"`
}

type BBox struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}
