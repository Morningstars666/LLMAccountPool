package handlers

import (
	"io"
	"llmaccountpool/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ImportResult struct {
	ExternalModelsCreated int      `json:"external_models_created"`
	ExternalModelsUpdated int      `json:"external_models_updated"`
	SourcesCreated        int      `json:"sources_created"`
	SourcesUpdated        int      `json:"sources_updated"`
	Skipped               []string `json:"skipped"`
	Errors                []string `json:"errors"`
}

func DownloadTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer f.Close()

	f.SetSheetName("Sheet1", "对外模型")
	f.SetCellValue("对外模型", "A1", "Name")
	f.SetCellValue("对外模型", "B1", "Model")
	f.SetCellValue("对外模型", "C1", "Strategy")

	f.NewSheet("上游模型")
	f.SetCellValue("上游模型", "A1", "ExternalModelName")
	f.SetCellValue("上游模型", "B1", "Name")
	f.SetCellValue("上游模型", "C1", "APIURL")
	f.SetCellValue("上游模型", "D1", "APIKey")
	f.SetCellValue("上游模型", "E1", "ModelName")
	f.SetCellValue("上游模型", "F1", "BillingMode")
	f.SetCellValue("上游模型", "G1", "LimitCount")
	f.SetCellValue("上游模型", "H1", "LimitTokens")
	f.SetCellValue("上游模型", "I1", "LimitResetInterval")
	f.SetCellValue("上游模型", "J1", "IsActive")

	f.SetColWidth("对外模型", "A", "A", 20)
	f.SetColWidth("对外模型", "B", "B", 20)
	f.SetColWidth("对外模型", "C", "C", 15)
	f.SetColWidth("上游模型", "A", "A", 20)
	f.SetColWidth("上游模型", "B", "B", 20)
	f.SetColWidth("上游模型", "C", "C", 40)
	f.SetColWidth("上游模型", "D", "D", 50)
	f.SetColWidth("上游模型", "E", "E", 20)
	f.SetColWidth("上游模型", "F", "F", 12)
	f.SetColWidth("上游模型", "G", "G", 12)
	f.SetColWidth("上游模型", "H", "H", 12)
	f.SetColWidth("上游模型", "I", "I", 18)
	f.SetColWidth("上游模型", "J", "J", 10)

	c.Header("Content-Disposition", "attachment; filename=models_template.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Stream(func(w io.Writer) bool {
		f.Write(w)
		return false
	})
}

func ImportModelsFromExcel(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse Excel file"})
		return
	}
	defer f.Close()

	result := &ImportResult{
		Skipped: []string{},
		Errors:  []string{},
	}

	externalModelMap := make(map[string]*models.ExternalModel)

	externalSheet := "对外模型"
	if sheetIndex, _ := f.GetSheetIndex(externalSheet); sheetIndex != -1 {
		rows, err := f.GetRows(externalSheet)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read ExternalModels sheet"})
			return
		}

		for i, row := range rows {
			if i == 0 {
				continue
			}
			if len(row) < 3 {
				result.Errors = append(result.Errors, "对外模型 sheet 第 "+strconv.Itoa(i+1)+" 行数据不完整")
				continue
			}

			name := strings.TrimSpace(row[0])
			model := strings.TrimSpace(row[1])
			strategy := strings.TrimSpace(row[2])
			if strategy == "" {
				strategy = "round_robin"
			}

			if name == "" || model == "" {
				result.Errors = append(result.Errors, "对外模型 sheet 第 "+strconv.Itoa(i+1)+" 行 name 或 model 为空")
				continue
			}

			var extModel models.ExternalModel
			if err := models.DB.Where("name = ?", name).First(&extModel).Error; err != nil {
				extModel = models.ExternalModel{
					Name:     name,
					Model:    model,
					Strategy: strategy,
				}
				if err := models.DB.Create(&extModel).Error; err != nil {
					result.Errors = append(result.Errors, "创建对外模型 "+name+" 失败: "+err.Error())
					continue
				}
				result.ExternalModelsCreated++
			} else {
				extModel.Model = model
				extModel.Strategy = strategy
				models.DB.Save(&extModel)
				result.ExternalModelsUpdated++
			}

			externalModelMap[name] = &extModel
		}
	}

	sourceSheet := "上游模型"
	if sheetIndex, _ := f.GetSheetIndex(sourceSheet); sheetIndex != -1 {
		rows, err := f.GetRows(sourceSheet)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read RequestSources sheet"})
			return
		}

		for i, row := range rows {
			if i == 0 {
				continue
			}
			if len(row) < 5 {
				result.Errors = append(result.Errors, "上游模型 sheet 第 "+strconv.Itoa(i+1)+" 行数据不完整")
				continue
			}

			externalModelName := strings.TrimSpace(row[0])
			name := strings.TrimSpace(row[1])
			apiURL := strings.TrimSpace(row[2])
			apiKey := strings.TrimSpace(row[3])
			modelName := strings.TrimSpace(row[4])

			if externalModelName == "" || name == "" || apiURL == "" || apiKey == "" || modelName == "" {
				result.Errors = append(result.Errors, "上游模型 sheet 第 "+strconv.Itoa(i+1)+" 行必填字段为空")
				continue
			}

			extModel, ok := externalModelMap[externalModelName]
			if !ok {
				var em models.ExternalModel
				if err := models.DB.Where("name = ?", externalModelName).First(&em).Error; err != nil {
					result.Skipped = append(result.Skipped, "上游模型 "+name+" 找不到对应对外模型: "+externalModelName)
					continue
				}
				extModel = &em
				externalModelMap[externalModelName] = &em
			}

			billingMode := "count"
			if len(row) > 5 && strings.TrimSpace(row[5]) != "" {
				billingMode = strings.TrimSpace(row[5])
			}

			limitCount := int64(0)
			if len(row) > 6 && strings.TrimSpace(row[6]) != "" {
				if v, err := strconv.ParseInt(strings.TrimSpace(row[6]), 10, 64); err == nil {
					limitCount = v
				}
			}

			limitTokens := int64(0)
			if len(row) > 7 && strings.TrimSpace(row[7]) != "" {
				if v, err := strconv.ParseInt(strings.TrimSpace(row[7]), 10, 64); err == nil {
					limitTokens = v
				}
			}

			limitResetInterval := int64(0)
			if len(row) > 8 && strings.TrimSpace(row[8]) != "" {
				if v, err := strconv.ParseInt(strings.TrimSpace(row[8]), 10, 64); err == nil {
					limitResetInterval = v
				}
			}

			isActive := true
			if len(row) > 9 && strings.TrimSpace(row[9]) != "" {
				isActive = strings.TrimSpace(row[9]) == "true" || strings.TrimSpace(row[9]) == "1" || strings.ToLower(strings.TrimSpace(row[9])) == "是"
			}

			var source models.RequestSource
			if err := models.DB.Where("external_model_id = ? AND name = ?", extModel.ID, name).First(&source).Error; err != nil {
				source = models.RequestSource{
					ExternalModelID:    extModel.ID,
					Name:               name,
					APIURL:             apiURL,
					APIKey:             apiKey,
					ModelName:          modelName,
					BillingMode:        billingMode,
					LimitCount:         limitCount,
					LimitTokens:        limitTokens,
					LimitResetInterval: limitResetInterval,
					IsActive:           isActive,
				}
				if err := models.DB.Create(&source).Error; err != nil {
					result.Errors = append(result.Errors, "创建上游模型 "+name+" 失败: "+err.Error())
					continue
				}
				result.SourcesCreated++
			} else {
				source.APIURL = apiURL
				source.APIKey = apiKey
				source.ModelName = modelName
				source.BillingMode = billingMode
				source.LimitCount = limitCount
				source.LimitTokens = limitTokens
				source.LimitResetInterval = limitResetInterval
				source.IsActive = isActive
				models.DB.Save(&source)
				result.SourcesUpdated++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "导入完成",
		"result":  result,
	})
}
