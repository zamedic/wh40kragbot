package embedd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/textsplitter"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"wh40k/internal/db"
	"wh40k/internal/llama"
)

const llamaImageDir = "llama-image-dir"
const llamaAPiKey = "llama-api-key"
const llamaTextDir = "llama-text-dir"
const pdfDownloadDirectory = "pdf-download-directory"

var PDFEmbedding = &cobra.Command{
	Use:   "pdf [yaml config file]",
	Run:   pdfEmbed,
	Short: "Download and embed PDFs",
}

func init() {

	PDFEmbedding.Flags().String(llamaAPiKey, "", "Llama API key")
	viper.BindPFlag(llamaAPiKey, PDFEmbedding.Flags().Lookup(llamaAPiKey))

	PDFEmbedding.Flags().String(llamaTextDir, "./text", "Directory to store text files")
	viper.BindPFlag(llamaTextDir, PDFEmbedding.Flags().Lookup(llamaTextDir))

	PDFEmbedding.Flags().String(llamaImageDir, "images", "Directory to store images")
	viper.BindPFlag(llamaImageDir, PDFEmbedding.Flags().Lookup(llamaImageDir))

	PDFEmbedding.Flags().String(pdfDownloadDirectory, "./pdf", "Directory to store downloaded PDFs")
	viper.BindPFlag(pdfDownloadDirectory, PDFEmbedding.Flags().Lookup(pdfDownloadDirectory))
}

var (
	llamaClient    *llama.Parse
	textDir        string
	imageDir       string
	llamaKey       string
	pdfDownloadDir string
)

func pdfEmbed(cmd *cobra.Command, args []string) {
	textDir = viper.GetString(llamaTextDir)
	imageDir = viper.GetString(llamaImageDir)
	llamaKey = viper.GetString(llamaAPiKey)
	pdfDownloadDir = viper.GetString(pdfDownloadDirectory)

	if err := os.MkdirAll(pdfDownloadDir, os.ModePerm); err != nil {
		zap.L().Panic("error creating pdf download directory", zap.Error(err))
	}

	llamaClient = llama.NewLlamaParse(textDir, imageDir, llamaKey)

	f, err := os.ReadFile(args[0])
	if err != nil {
		zap.L().Panic("error reading file", zap.Error(err))
	}
	config := pdfEmbeddingConfig{}
	err = yaml.Unmarshal(f, &config)
	if err != nil {
		zap.L().Panic("error unmarshalling yaml", zap.Error(err))
	}
	zap.L().Debug("pdf embedding config", zap.Any("config", config))

	for _, rule := range config.Rules {
		if err := process(cmd.Context(), rule); err != nil {
			zap.L().Error("error processing rule", zap.Error(err))
		}
	}
	for _, index := range config.Indexes {
		if err := process(cmd.Context(), index); err != nil {
			zap.L().Error("error processing index", zap.Error(err))
		}
	}

}

func process(ctx context.Context, index PdfEmbedding) error {
	//Download the PDF
	zap.L().Debug("downloading PDF", zap.String("url", index.Url))
	pdfPath, err := downloadFile(index)
	if err != nil {
		return err
	}

	//OCR the PDF
	zap.L().Debug("parsing PDF", zap.String("path", pdfPath))
	pages, err := llamaClient.Parse(ctx, pdfPath)
	if err != nil {
		zap.L().Error("error parsing PDF", zap.Error(err))
		return err
	}

	//Embed the text
	zap.L().Debug("embedding text")
	llm, err := ollama.New(ollama.WithModel("nomic-embed-text"))
	if err != nil {
		zap.L().Error("error creating LLM", zap.Error(err))
		return err
	}

	splitter := textsplitter.NewRecursiveCharacter()

	for _, page := range pages.Pages {
		texts, err := splitter.SplitText(page.Md)
		if err != nil {
			zap.L().Error("error splitting text", zap.Error(err))
			return err
		}
		tokens, err := llm.CreateEmbedding(ctx, texts)
		if err != nil {
			zap.L().Error("error creating embedding", zap.Error(err))
			return err
		}

		embeddings := make([]*db.Embedding, len(tokens))

		for x, token := range tokens {
			embeddings[x] = &db.Embedding{
				Vector:   token,
				Page:     page.Page,
				Document: index.Title,
				Index:    x,
			}
		}

		_, err = db.MongoViper(ctx).Database("wh40k").Collection("embeddings").InsertMany(ctx, []interface{}{embeddings})
		if err != nil {
			zap.L().Error("error inserting embeddings", zap.Error(err))
			return err
		}

	}

	return nil

}

func downloadFile(index PdfEmbedding) (string, error) {
	response, err := http.DefaultClient.Get(index.Url)
	if err != nil {
		zap.L().Error("error downloading PDF", zap.Error(err))
		return "", err
	}
	defer response.Body.Close()

	//save the PDF
	pdfPath := filepath.Join(pdfDownloadDir, index.Title+".pdf")
	pdfFile, err := os.Create(pdfPath)
	if err != nil {
		zap.L().Error("error creating PDF", zap.Error(err))
		return "", err
	}
	_, err = io.Copy(pdfFile, response.Body)
	if err != nil {
		zap.L().Error("error saving PDF", zap.Error(err))
		return "", err
	}
	err = pdfFile.Close()
	if err != nil {
		zap.L().Error("error closing PDF", zap.Error(err))
		return "", err
	}
	return pdfPath, nil
}
