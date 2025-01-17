package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"wh40k/cmd/embedd"
)

var rootCmd = &cobra.Command{
	Use:   "whbot",
	Short: "Warhammer 40k RAG bot",
}

func Execute() error {
	return rootCmd.Execute()
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in home directory with name ".cobra" (without extension).
	viper.AddConfigPath(home)
	viper.AddConfigPath(".")

	viper.SetConfigType("yaml")
	viper.SetConfigName(".whbot.yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		zap.L().Fatal("Error reading config file", zap.Error(err))
	}
}

func init() {
	setupZap()
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("mongo-url", "mongodb://localhost:27017", "MongoDB URL")
	viper.BindPFlag("mongo-url", rootCmd.PersistentFlags().Lookup("mongo-url"))

	rootCmd.AddCommand(embedd.PDFEmbedding)

}

func setupZap() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}
