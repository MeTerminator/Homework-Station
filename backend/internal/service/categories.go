package service

import (
	"strings"

	"gorm.io/gorm"

	"homework-station/internal/model"
)

type CategoryService struct {
	db    *gorm.DB
	files *FileService
}

func NewCategoryService(db *gorm.DB, files *FileService) *CategoryService {
	return &CategoryService{db: db, files: files}
}
func (svc *CategoryService) List() ([]model.Category, error) {
	items := []model.Category{}
	err := svc.db.Order("sort asc, id asc").Find(&items).Error
	return items, err
}
func (svc *CategoryService) Save(v model.Category) (model.Category, error) {
	svc.files.MutationMu.Lock()
	defer svc.files.MutationMu.Unlock()
	if strings.TrimSpace(v.Name) == "" || len(v.Name) > 100 {
		return model.Category{}, Reject(400, "分类名称必填且不超过 100 字节")
	}
	if v.ParentID != nil {
		seen := map[uint]bool{v.ID: true}
		id := *v.ParentID
		for {
			if seen[id] {
				return model.Category{}, Reject(400, "分类不能形成循环")
			}
			seen[id] = true
			var parent model.Category
			if svc.db.First(&parent, id).Error != nil {
				return model.Category{}, Reject(400, "父分类不存在")
			}
			if parent.ParentID == nil {
				break
			}
			id = *parent.ParentID
		}
	}

	if v.ID == 0 {
		if svc.db.Create(&v).Error != nil {
			return model.Category{}, Reject(500, "创建失败")
		}
	} else {
		var old model.Category
		if svc.db.First(&old, v.ID).Error != nil {
			return model.Category{}, Reject(404, "分类不存在")
		}
		if svc.db.Model(&old).Select("Name", "ParentID", "Sort").Updates(v).Error != nil {
			return model.Category{}, Reject(500, "保存失败")
		}
	}
	return v, nil
}
func (svc *CategoryService) Delete(id uint) error {
	svc.files.MutationMu.Lock()
	defer svc.files.MutationMu.Unlock()
	var n int64
	svc.db.Model(&model.Category{}).Where("parent_id = ?", id).Count(&n)
	if n > 0 {
		return Reject(409, "请先移动或删除子分类")
	}
	svc.db.Model(&model.Assignment{}).Where("category_id = ?", id).Count(&n)
	if n > 0 {
		return Reject(409, "分类下仍有作业，请先移动作业")
	}
	if svc.db.Delete(&model.Category{}, id).RowsAffected == 0 {
		return Reject(404, "分类不存在")
	}
	return nil
}
