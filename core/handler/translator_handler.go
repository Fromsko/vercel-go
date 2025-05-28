package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	translator "github.com/package-register/go-genius/trans"
	"github.com/spf13/viper"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Warning: .env file not found, using system environment variables\n")
	}
	viper.AutomaticEnv()
}

func WrapperTranslator(from, to string) translator.Translator {
	appID := viper.GetString("TRANSLATOR_APP_ID")
	secret := viper.GetString("TRANSLATOR_SECRET")
	apiKey := viper.GetString("TRANSLATOR_API_KEY")

	return translator.New(
		translator.WithAppID(appID),
		translator.WithSecret(secret),
		translator.WithAPIKey(apiKey),
		translator.WithToLang(to),
		translator.WithFromLang(from),
	)
}

// TranslateRequest 请求结构体
type TranslateRequest struct {
	FromLang string   `json:"fromLang"`
	ToLang   string   `json:"toLang"`
	Texts    []string `json:"texts" binding:"required"`
}

// TranslateResponse 响应结构体
type TranslateResponse struct {
	Results []map[string]string `json:"results"`
}

// TranslateHandler 处理翻译请求
func TranslateHandler(c *gin.Context) {
	var req TranslateRequest

	// 解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 检查是否有文本需要翻译
	if len(req.Texts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No texts provided"})
		return
	}

	transWrapper := WrapperTranslator(req.FromLang, req.ToLang)
	results := make([]map[string]string, 0, len(req.Texts))

	// 逐个翻译（动态传入 fromLang 和 toLang）
	for _, text := range req.Texts {
		translatedText, err := transWrapper.TranslateWithResult(text)
		if err != nil {
			results = append(results, map[string]string{
				"original":   text,
				"translated": "Translation failed",
				"error":      err.Error(),
			})
			continue
		}

		results = append(results, map[string]string{
			"original":   translatedText.Source,
			"translated": translatedText.Target,
		})
	}

	// 返回结果
	c.JSON(http.StatusOK, TranslateResponse{
		Results: results,
	})
}
