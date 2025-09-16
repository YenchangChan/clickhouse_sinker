package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func MergeConfig(config interface{}, jsonData []byte) error {
	if len(jsonData) == 0 {
		return nil
	}
	// 解析JSON数据到map
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 使用反射处理结构体
	val := reflect.ValueOf(config)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	return mergeStruct(val.Elem(), data, "")
}

// mergeStruct 递归处理结构体字段
func mergeStruct(v reflect.Value, data map[string]interface{}, basePath string) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// 获取merge标签值
		tag := field.Tag.Get("merge")
		if tag == "" {
			// 无标签字段：如果是结构体则递归处理
			if fieldVal.Kind() == reflect.Struct {
				if err := mergeStruct(fieldVal, data, basePath); err != nil {
					return err
				}
			}
			continue
		}

		// 构建完整路径
		fullPath := tag
		if basePath != "" {
			fullPath = basePath + "." + tag
		}

		// 从JSON数据中获取值
		jsonValue, found := getValueByPath(data, fullPath)
		if !found {
			continue
		}

		if !isFieldValueZero(fieldVal) && fieldVal.Kind() != reflect.Map {
			// 如果字段值不为零值，则跳过, 注意，map有特殊处理
			continue
		}

		// 检查并设置字段值
		if err := setFieldValue(fieldVal, jsonValue); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
	}

	return nil
}

// getValueByPath 通过点分隔路径从map中获取值
func getValueByPath(data map[string]interface{}, path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	current := data

	for i, key := range keys {
		value, ok := current[key]
		if !ok {
			return nil, false
		}

		// 如果是最后一级键，直接返回值
		if i == len(keys)-1 {
			return value, true
		}

		// 继续深入下一层map
		next, ok := value.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current = next
	}

	return nil, false
}

// setFieldValue 设置字段值并进行类型检查
func setFieldValue(field reflect.Value, value interface{}) error {
	// 获取字段实际类型（解引用指针）
	fieldType := field.Type()
	if fieldType.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(fieldType.Elem()))
		}
		field = field.Elem()
		fieldType = field.Type()
	}

	// 特殊处理：多维切片（如[][]string）
	if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Slice {
		return setMultiDimensionalSlice(field, value)
	}

	// 特殊处理：map[string]string 类型
	if fieldType.Kind() == reflect.Map &&
		fieldType.Key().Kind() == reflect.String &&
		fieldType.Elem().Kind() == reflect.String {
		return setStringMap(field, value)
	}

	// 检查JSON值类型是否匹配
	jsonType := reflect.TypeOf(value)
	if !jsonType.ConvertibleTo(fieldType) {
		return fmt.Errorf("type mismatch: expected %s, got %s", fieldType, jsonType)
	}

	// 常规类型设置
	field.Set(reflect.ValueOf(value).Convert(fieldType))
	return nil
}

// setMultiDimensionalSlice 处理多维切片类型
func setMultiDimensionalSlice(field reflect.Value, value interface{}) error {
	// 验证JSON值类型
	jsonSlice, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected slice, got %T", value)
	}

	// 创建目标切片
	elemType := field.Type().Elem()
	result := reflect.MakeSlice(field.Type(), len(jsonSlice), len(jsonSlice))

	for i, item := range jsonSlice {
		// 验证内层切片类型
		innerSlice, ok := item.([]interface{})
		if !ok {
			return fmt.Errorf("expected inner slice, got %T", item)
		}

		// 创建内层切片
		innerResult := reflect.MakeSlice(elemType, len(innerSlice), len(innerSlice))
		for j, val := range innerSlice {
			innerResult.Index(j).Set(reflect.ValueOf(val))
		}

		result.Index(i).Set(innerResult)
	}

	field.Set(result)
	return nil
}

// setStringMap 处理 map[string]string 类型，实现新的合并逻辑
func setStringMap(field reflect.Value, value interface{}) error {
	// 验证JSON值类型
	jsonMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", value)
	}

	fmt.Printf("jsonMap: %#v\n", jsonMap)

	// 获取现有的map或创建新的map（如果字段为空）
	var existingMap map[string]string
	if !field.IsNil() {
		existingMap = field.Interface().(map[string]string)
	} else {
		existingMap = make(map[string]string)
	}
	fmt.Printf("existingMap: %#v\n", existingMap)

	// 只添加JSON中新增的键值对，不修改已有的值
	for k, v := range jsonMap {
		// 只有当键不存在时才添加
		if _, exists := existingMap[k]; !exists {
			existingMap[k] = fmt.Sprintf("%v", v)
		}
	}

	// 将更新后的map设置回字段
	field.Set(reflect.ValueOf(existingMap))
	return nil
}

// isFieldValueZero 检查字段值是否为零值
func isFieldValueZero(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan:
		return field.IsNil()
	case reflect.Struct:
		// 对于结构体，检查是否所有字段都为零值
		return reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface())
	default:
		return reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface())
	}
}
