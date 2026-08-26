package controllers

import (
	"base-backend/models"
	"base-backend/services"
	"base-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
