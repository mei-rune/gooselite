package gooselite

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

// common routines

func writeTemplateToFile(path string, t *template.Template, data interface{}) (string, error) {
	f, e := os.Create(path)
	if e != nil {
		return "", e
	}
	defer f.Close()

	e = t.Execute(f, data)
	if e != nil {
		return "", e
	}

	return f.Name(), nil
}

// TableAlreadyExistKeys 表已存在的错误关键词
var TableAlreadyExistKeys = []string{"already exists", "已存在", "已经存在"}

func IsTableAlreadyExists(err error) bool {
	if errors.Is(err, ErrTableDoesAlreadyExist) {
		return true
	}
	for _, key := range TableAlreadyExistKeys {
		if strings.Contains(err.Error(), key) {
			return true
		}
	}
	return false
}

// TableNotExistKeys 表不存在的错误关键词
var TableNotExistKeys = []string{"does not exist", "doesn't exist", "不存在"}

func IsTableNotExists(err error) bool {
	if errors.Is(err, ErrTableDoesNotExist) {
		return true
	}
	for _, key := range TableNotExistKeys {
		if strings.Contains(err.Error(), key) {
			return true
		}
	}
	return false
}

var validPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`) // 小写字母+下划线

// 表名校验（正则验证）
func ValidateTableName(table string) error {
	if !validPattern.MatchString(table) {
		return fmt.Errorf("表名 %q 不符合命名规范", table)
	}
	return nil
}

// ValidateTableNames validates multiple table names.
func ValidateTableNames(names ...string) error {
	for _, name := range names {
		if err := ValidateTableName(name); err != nil {
			return err
		}
	}
	return nil
}
