package embedd

type pdfEmbeddingConfig struct {
	Rules   []PdfEmbedding `yaml:"rules"`
	Indexes []PdfEmbedding `yaml:"indexes"`
}

type PdfEmbedding struct {
	Title string `yaml:"title"`
	Url   string `yaml:"url"`
}
