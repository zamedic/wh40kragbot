package embedd

type pdfEmbeddingConfig struct {
	Rules   []pdfEmbedding `yaml:"rules"`
	Indexes []pdfEmbedding `yaml:"indexes"`
}

type pdfEmbedding struct {
	Title string `yaml:"title"`
	Url   string `yaml:"url"`
}
