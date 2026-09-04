package controllers

import (
	"base-backend/models"
	"base-backend/services"
	"base-backend/utils"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// dictExportContentType xlsx 的 MIME 类型，导出/导入共用
const dictExportContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

// dictExportColumns 字典项导入导出一致的列头（D6：模板与导出约定相同，避免列名错位）
var dictExportColumns = []interface{}{"label", "value", "sort", "status", "remark"}

// writeDictItemsSheet 把一批字典项写入指定 sheet：首行表头，其后按行写 label/value/sort/status/remark
func writeDictItemsSheet(f *excelize.File, sheet string, items []models.DictItem) {
	f.SetSheetRow(sheet, "A1", &dictExportColumns)
	for i, it := range items {
		row := []interface{}{it.Label, it.Value, it.Sort, it.Status, it.Remark}
		if cell, err := excelize.CoordinatesToCellName(1, i+2); err == nil {
			f.SetSheetRow(sheet, cell, &row)
		}
	}
}

// sanitizeSheetName 清洗 sheet 名：去非法字符、截断到 31 字符；空则返回空串由调用方兜底
func sanitizeSheetName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) > 31 {
		s = s[:31]
	}
	replacer := strings.NewReplacer("\\", "_", "/", "_", "?", "_", "*", "_", ":", "_", "[", "_", "]", "_", "'", "_")
	return replacer.Replace(s)
}

// uniqueSheetName 生成不重复的 sheet 名：优先 type.name，冲突/非法时回退 code，
// 仍冲突则追加序号（设计 D3 全量导出 sheet 命名约束）
func uniqueSheetName(name, code string, used map[string]int) string {
	base := sanitizeSheetName(name)
	if base == "" {
		base = sanitizeSheetName(code)
	}
	if base == "" {
		base = "字典"
	}
	candidate := base
	for i := 2; used[candidate] > 0; i++ {
		suffix := fmt.Sprintf("_%d", i)
		candidate = base
		if len(candidate) > 31-len(suffix) {
			candidate = candidate[:31-len(suffix)]
		}
		candidate += suffix
	}
	used[candidate] = 1
	return candidate
}

// GetDictTypeList 字典类型列表
func GetDictTypeList(c *gin.Context) {
	page, pageSize := utils.GetPageParams(c.Query("page"), c.Query("page_size"))
	keyword := c.Query("keyword")

	types, total, err := services.GetDictTypeList(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, types, total, page, pageSize)
}

// GetDictType 字典类型详情
func GetDictType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的字典类型ID")
		return
	}

	t, err := services.GetDictType(uint(id))
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, t)
}

// CreateDictType 创建字典类型
func CreateDictType(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
		Sort        int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	t := &models.DictType{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      1,
	}
	if err := services.CreateDictType(t); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, t)
}

// UpdateDictType 更新字典类型
func UpdateDictType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的字典类型ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
		Sort        *int   `json:"sort"`
		Status      *int   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := services.UpdateDictType(uint(id), updates); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// DeleteDictType 删除字典类型
func DeleteDictType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的字典类型ID")
		return
	}

	if err := services.DeleteDictType(uint(id)); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetDictItemList 字典项列表
func GetDictItemList(c *gin.Context) {
	page, pageSize := utils.GetPageParams(c.Query("page"), c.Query("page_size"))
	dictTypeID, _ := strconv.Atoi(c.Query("dict_type_id"))

	items, total, err := services.GetDictItemList(page, pageSize, uint(dictTypeID))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessPage(c, items, total, page, pageSize)
}

// GetDictItemsByCode 根据编码获取字典项
func GetDictItemsByCode(c *gin.Context) {
	code := c.Param("code")
	items, err := services.GetDictItemsByCode(code)
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, items)
}

// CreateDictItem 创建字典项
func CreateDictItem(c *gin.Context) {
	var req struct {
		DictTypeID uint   `json:"dict_type_id" binding:"required"`
		Label      string `json:"label" binding:"required"`
		Value      string `json:"value" binding:"required"`
		Sort       int    `json:"sort"`
		Remark     string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	item := &models.DictItem{
		DictTypeID: req.DictTypeID,
		Label:      req.Label,
		Value:      req.Value,
		Sort:       req.Sort,
		Status:     1,
		Remark:     req.Remark,
	}
	if err := services.CreateDictItem(item); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, item)
}

// UpdateDictItem 更新字典项
func UpdateDictItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的字典项ID")
		return
	}

	var req struct {
		Label  string `json:"label"`
		Value  string `json:"value"`
		Sort   *int   `json:"sort"`
		Status *int   `json:"status"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Label != "" {
		updates["label"] = req.Label
	}
	if req.Value != "" {
		updates["value"] = req.Value
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	if err := services.UpdateDictItem(uint(id), updates); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// DeleteDictItem 删除字典项
func DeleteDictItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的字典项ID")
		return
	}

	if err := services.DeleteDictItem(uint(id)); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
}

// ExportDictItems 单类型字典项导出（也用于生成导入模板）：生成 xlsx 并触发下载
func ExportDictItems(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的字典类型ID")
		return
	}
	t, err := services.GetDictType(uint(id))
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	items, err := services.GetDictItemsAll(uint(id))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "字典项"
	f.SetSheetName("Sheet1", sheet)
	writeDictItemsSheet(f, sheet, items)

	serveDictXLSX(c, t.Name+".xlsx", f)
}

// ExportAllDictTypes 全量字典导出：每个类型一个 sheet，sheet 名取类型中文名，冲突/非法回退 code
func ExportAllDictTypes(c *gin.Context) {
	types, err := services.GetAllDictTypesWithItems()
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	f.DeleteSheet("Sheet1")

	used := make(map[string]int)
	for _, t := range types {
		sheet := uniqueSheetName(t.Name, t.Code, used)
		if _, err := f.NewSheet(sheet); err != nil {
			utils.Fail(c, 500, "创建工作表失败")
			return
		}
		writeDictItemsSheet(f, sheet, t.Items)
	}
	// 空库时兜底保留一个含表头的 sheet
	if len(types) == 0 {
		f.NewSheet("字典")
		writeDictItemsSheet(f, "字典", nil)
	}

	serveDictXLSX(c, "全量字典.xlsx", f)
}

// serveDictXLSX 把 excelize 文件写入缓冲区并以下载形式返回
func serveDictXLSX(c *gin.Context, filename string, f *excelize.File) {
	buf, err := f.WriteToBuffer()
	if err != nil {
		utils.Fail(c, 500, "生成文件失败")
		return
	}
	// 中文文件名用 RFC5987 编码，避免浏览器乱码或被截断
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(filename))
	c.Data(200, dictExportContentType, buf.Bytes())
}

// ImportDictItems 上传 xlsx 并按 value 覆盖合并导入当前类型字典项
func ImportDictItems(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "无效的字典类型ID")
		return
	}
	typeID := uint(id)
	if _, err := services.GetDictType(typeID); err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}

	header, err := c.FormFile("file")
	if err != nil {
		utils.Fail(c, 400, "请上传文件")
		return
	}
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		utils.Fail(c, 400, "仅支持 .xlsx 格式文件")
		return
	}
	src, err := header.Open()
	if err != nil {
		utils.Fail(c, 500, "读取文件失败")
		return
	}
	defer src.Close()

	xlsxFile, err := excelize.OpenReader(src)
	if err != nil {
		utils.Fail(c, 400, "无法解析文件，请使用导入模板")
		return
	}
	defer xlsxFile.Close()

	rows, err := parseDictImportRows(xlsxFile)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	newCount, updatedCount, err := services.ImportDictItems(typeID, rows)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"new": newCount, "updated": updatedCount})
}

// parseDictImportRows 读取首个 sheet 的数据行，跳过表头；label/value 缺失的行在 service 层跳过
func parseDictImportRows(f *excelize.File) ([]services.DictItemImportRow, error) {
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("文件中没有工作表")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, errors.New("读取工作表失败")
	}

	parseStatus := func(s string) int {
		s = strings.TrimSpace(s)
		if s == "" {
			return 1 // 缺省启用（与模型 status 默认一致）
		}
		n, _ := strconv.Atoi(s)
		return n
	}
	parseInt := func(s string) int {
		n, _ := strconv.Atoi(strings.TrimSpace(s))
		return n
	}

	var out []services.DictItemImportRow
	for i, row := range rows {
		if i == 0 || len(row) == 0 {
			continue // 首行为表头
		}
		cell := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		r := services.DictItemImportRow{
			Label:  cell(0),
			Value:  cell(1),
			Sort:   parseInt(cell(2)),
			Status: parseStatus(cell(3)),
			Remark: cell(4),
		}
		// 全空行跳过
		if r.Label == "" && r.Value == "" && r.Remark == "" && r.Sort == 0 && r.Status == 1 {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}
