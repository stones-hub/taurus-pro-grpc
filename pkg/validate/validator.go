// Copyright (c) 2025 Taurus Team. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: yelei
// Email: 61647649@qq.com
// Date: 2025-06-13

package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// 核心验证器
var Core = New()

// Validator 包装了validator.v10，提供了中文错误消息
type Validator struct {
	validate *validator.Validate // 验证器
	trans    ut.Translator       // 翻译器
}

// 错误类型
type ValidationError struct {
	Field   string // 字段名
	Tag     string // 验证标签
	Value   any    // 字段值
	Message string // 错误消息
	Param   string // 验证参数
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	return e.Message
}

// ValidationErrors 表示多个验证错误
type ValidationErrors []*ValidationError

// Error 实现error接口
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, err := range e {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// New 创建一个新的Validator实例
func New() *Validator {
	en := en.New()        // 创建英文翻译器
	zh := zh.New()        // 创建中文翻译器
	uni := ut.New(en, zh) // 创建通用翻译器，优先使用中文，使用英文作为回退语言
	// 获取中文翻译器, 如果你想获取其他语言的翻译器，请使用uni.GetTranslator("其他语言")
	trans, _ := uni.GetTranslator("zh")

	// 创建验证器
	validate := validator.New()

	// 验证器注册翻译器
	_ = zh_translations.RegisterDefaultTranslations(validate, trans)

	// 优先获取tag中json作为错误消息提示的字段名称，如果json不存在，则使用字段名
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	return &Validator{
		validate: validate, // 验证器
		trans:    trans,    // 翻译器
	}
}

// Validate 验证结构体
func (v *Validator) ValidateStruct(value interface{}) (ValidationErrors, error) {
	err := v.validate.Struct(value)
	if err == nil {
		return nil, nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, fmt.Errorf("无效的验证类型: %v", err)
	}

	// 将错误转换为我们自己的错误类型
	result := make(ValidationErrors, 0, len(validationErrors))
	for _, e := range validationErrors {
		result = append(result, &ValidationError{
			Field:   e.Field(),
			Tag:     e.Tag(),
			Value:   e.Value(),
			Message: e.Translate(v.trans),
			Param:   e.Param(),
		})
	}

	return result, nil
}

// ValidateVar 验证单个变量, tag 参数是验证规则的字符串
// validators.ValidateVar(email, "email")  // 验证邮箱格式
// validators.ValidateVar(password, "min=8")  // 验证最小长度
// validators.ValidateVar(age, "gte=18,lte=120")  // 验证数值范围
func (v *Validator) ValidateVar(field interface{}, tag string) (ValidationErrors, error) {
	err := v.validate.Var(field, tag)
	if err == nil {
		return nil, nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, fmt.Errorf("无效的验证类型: %v", err)
	}

	// 将错误转换为我们自己的错误类型
	result := make(ValidationErrors, 0, len(validationErrors))
	for _, e := range validationErrors {
		result = append(result, &ValidationError{
			Field:   "",
			Tag:     e.Tag(),
			Value:   e.Value(),
			Message: e.Translate(v.trans),
			Param:   e.Param(),
		})
	}

	return result, nil
}

// ValidateVarWithValue 验证变量与另一个变量的关系
func (v *Validator) ValidateVarWithValue(field, other interface{}, tag string) (ValidationErrors, error) {
	err := v.validate.VarWithValue(field, other, tag)
	if err == nil {
		return nil, nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, fmt.Errorf("无效的验证类型: %v", err)
	}

	// 将错误转换为我们自己的错误类型
	result := make(ValidationErrors, 0, len(validationErrors))
	for _, e := range validationErrors {
		result = append(result, &ValidationError{
			Field:   "",
			Tag:     e.Tag(),
			Value:   e.Value(),
			Message: e.Translate(v.trans),
			Param:   e.Param(),
		})
	}

	return result, nil
}

// ValidateMap 验证映射
func (v *Validator) ValidateMap(m interface{}) (ValidationErrors, error) {
	val := reflect.ValueOf(m)
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("输入不是映射类型")
	}

	var allErrors ValidationErrors

	// 遍历映射的键值对
	for _, key := range val.MapKeys() {
		value := val.MapIndex(key)

		// 如果值是结构体或结构体指针，尝试验证它
		if value.Kind() == reflect.Struct ||
			(value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Struct) {
			errors, err := v.ValidateStruct(value.Interface())
			if err != nil {
				return nil, err
			}
			if len(errors) > 0 {
				// 添加映射键前缀到错误字段
				keyStr := fmt.Sprintf("%v", key.Interface())
				for _, e := range errors {
					e.Field = keyStr + "." + e.Field
					allErrors = append(allErrors, e)
				}
			}
		}
	}

	if len(allErrors) > 0 {
		return allErrors, nil
	}

	return nil, nil
}

// ValidateSlice 验证切片
func (v *Validator) ValidateSlice(slice interface{}) (ValidationErrors, error) {
	val := reflect.ValueOf(slice)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("输入不是切片或数组类型")
	}

	var allErrors ValidationErrors

	// 遍历切片元素
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i)

		// 如果元素是结构体或结构体指针，尝试验证它
		if item.Kind() == reflect.Struct ||
			(item.Kind() == reflect.Ptr && item.Elem().Kind() == reflect.Struct) {
			errors, err := v.ValidateStruct(item.Interface())
			if err != nil {
				return nil, err
			}
			if len(errors) > 0 {
				// 添加索引前缀到错误字段
				for _, e := range errors {
					e.Field = fmt.Sprintf("[%d].%s", i, e.Field)
					allErrors = append(allErrors, e)
				}
			}
		}
	}

	if len(allErrors) > 0 {
		return allErrors, nil
	}

	return nil, nil
}

// RegisterCustomValidation 注册自定义验证, tag 参数是验证规则的名称, fn 参数是验证函数, errMsg 参数是验证失败时的错误消息
func (v *Validator) RegisterCustomValidation(tag string, fn validator.Func, errMsg string) error {
	err := v.validate.RegisterValidation(tag, fn)
	if err != nil {
		return err
	}

	// 注册自定义验证的错误消息翻译
	return v.validate.RegisterTranslation(tag, v.trans, func(ut ut.Translator) error {
		return ut.Add(tag, errMsg, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field())
		return t
	})
}

// Engine 返回底层的validator引擎
func (v *Validator) Engine() *validator.Validate {
	return v.validate
}

// ValidateStruct 是一个便捷函数，用于验证结构体
// 如果验证失败，返回错误信息
func ValidateStruct(value interface{}) error {
	errors, err := Core.ValidateStruct(value)
	if err != nil {
		return err
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// ValidateVar 是一个便捷函数，用于验证单个变量
func ValidateVar(value interface{}, tag string) error {
	errors, err := Core.ValidateVar(value, tag)
	if err != nil {
		return err
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// ValidateVarWithValue 是一个便捷函数，用于验证变量与另一个变量的关系
func ValidateVarWithValue(value, other interface{}, tag string) error {
	errors, err := Core.ValidateVarWithValue(value, other, tag)
	if err != nil {
		return err
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// ValidateMap 是一个便捷函数，用于验证映射
func ValidateMap(m interface{}) error {
	errors, err := Core.ValidateMap(m)
	if err != nil {
		return err
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// ValidateSlice 是一个便捷函数，用于验证切片
func ValidateSlice(slice interface{}) error {
	errors, err := Core.ValidateSlice(slice)
	if err != nil {
		return err
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// GetFieldErrors 返回字段错误的映射，便于处理特定字段错误
func GetFieldErrors(errors ValidationErrors) map[string]string {
	result := make(map[string]string)
	for _, err := range errors {
		result[err.Field] = err.Message
	}
	return result
}
