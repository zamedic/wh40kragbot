package embedd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/textsplitter"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"wh40k/internal/llama"
)

const llamaImageDir = "llama-image-dir"
const llamaAPiKey = "llama-api-key"
const llamaTextDir = "llama-text-dir"
const pdfDownloadDirectory = "pdf-download-directory"

var PDFEmbedding = cobra.Command{
	Use:   "pdf [yaml config file]",
	Run:   pdfEmbed,
	Short: "Download and embed PDFs",
}

func init() {

	PDFEmbedding.Flags().String(llamaAPiKey, "", "Llama API key")
	viper.BindPFlag(llamaAPiKey, PDFEmbedding.Flags().Lookup(llamaAPiKey))

	PDFEmbedding.Flags().String(llamaTextDir, "./text", "Directory to store text files")
	viper.BindPFlag(llamaTextDir, PDFEmbedding.Flags().Lookup(llamaTextDir))

	PDFEmbedding.Flags().String(llamaImageDir, "./images", "Directory to store images")
	viper.BindPFlag(llamaImageDir, PDFEmbedding.Flags().Lookup(llamaImageDir))

	PDFEmbedding.Flags().String(pdfDownloadDirectory, "./pdf", "Directory to store downloaded PDFs")
	viper.BindPFlag(pdfDownloadDirectory, PDFEmbedding.Flags().Lookup(pdfDownloadDirectory))
}

var llamaClient *llama.Parse

func pdfEmbed(cmd *cobra.Command, args []string) {
	llamaClient = llama.NewLlamaParse()

	f, err := os.ReadFile(args[0])
	if err != nil {
		zap.L().Panic("error reading file", zap.Error(err))
	}
	config := pdfEmbeddingConfig{}
	err = yaml.Unmarshal(f, config)
	if err != nil {
		zap.L().Panic("error unmarshalling yaml", zap.Error(err))
	}
	zap.L().Debug("pdf embedding config", zap.Any("config", config))

	for _, rule := range config.Rules {
		process(cmd.Context(), rule)
	}
	for _, index := range config.Indexes {
		process(cmd.Context(), index)
	}

}

func process(ctx context.Context, index pdfEmbedding) error {
	//Download the PDF
	response, err := http.DefaultClient.Get(index.Url)
	if err != nil {
		zap.L().Error("error downloading PDF", zap.Error(err))
		return err
	}
	defer response.Body.Close()

	//save the PDF
	pdfPath := filepath.Join(viper.GetString(pdfDownloadDirectory), index.Title+".pdf")
	pdfFile := os.NewFile(0, pdfPath)
	_, err = io.Copy(pdfFile, response.Body)
	if err != nil {
		zap.L().Error("error saving PDF", zap.Error(err))
		return err
	}
	err = pdfFile.Close()
	if err != nil {
		zap.L().Error("error closing PDF", zap.Error(err))
		return err
	}

	//OCR the PDF
	err = llamaClient.Parse(ctx, pdfPath)
	if err != nil {
		zap.L().Error("error parsing PDF", zap.Error(err))
		return err
	}

	s := textsplitter.NewMarkdownTextSplitter()

}
