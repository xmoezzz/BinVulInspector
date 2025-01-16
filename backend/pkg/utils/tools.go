package utils

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

func Contains[T comparable](slice []T, elem T) bool {
	for i := range slice {
		if slice[i] == elem {
			return true
		}
	}
	return false
}

func ContainsFunc[T any](slice []T, f func(T) bool) bool {
	for i := range slice {
		if f(slice[i]) {
			return true
		}
	}
	return false
}

func ConvertToInterfaceSlice[T any](slice []T) []interface{} {
	// 创建接口切片
	result := make([]interface{}, len(slice))

	// 将切片的元素逐个转换为接口类型
	for i := range slice {
		result[i] = slice[i]
	}

	return result
}

func ToLowerSlice(in []string) []string {
	r := make([]string, len(in))
	for i, str := range in {
		r[i] = strings.ToLower(str)
	}
	return r
}

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateToken() string {
	uniqueKey := GenerateUUID()
	return base64.StdEncoding.EncodeToString([]byte(uniqueKey))
}

func Intersect[T comparable](s1, s2 []T) (res []T) {
	m := make(map[T]struct{})
	for _, v := range s1 {
		m[v] = struct{}{}
	}

	for _, v := range s2 {
		if _, ok := m[v]; ok {
			res = append(res, v)
		}
	}
	return res
}

func MD5(input string) string {
	hash := md5.New()
	hash.Write([]byte(input))
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func UniqueSlice[T comparable](items []T) []T {
	m := make(map[T]struct{})

	result := make([]T, 0)
	for _, item := range items {
		if _, ok := m[item]; !ok {
			m[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// IsInSubset 判断共集
func IsInSubset[T comparable](array, subset []T) bool {
	subsetSet := make(map[T]struct{})
	for _, item := range subset {
		subsetSet[item] = struct{}{}
	}
	for _, item := range array {
		if _, exists := subsetSet[item]; exists {
			delete(subsetSet, item)
			if len(subsetSet) == 0 {
				return true
			}
		}
	}
	return len(subsetSet) == 0
}

// SubsetList 取共集且去重
func SubsetList[T comparable](array, subset []T) (res []T) {
	subsetSet := make(map[T]struct{})
	for _, item := range subset {
		subsetSet[item] = struct{}{}
	}
	for _, item := range array {
		if _, exists := subsetSet[item]; exists {
			res = append(res, item)
			delete(subsetSet, item)
			if len(subsetSet) == 0 {
				return
			}
		}
	}
	return res
}

func CloneViaJson(input interface{}, output interface{}) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, output)
}

func NotNull[T any](s []T) []T {
	if len(s) == 0 {
		return make([]T, 0)
	}
	return s
}

// NotNullMap 检查map是否为nil，如果是，则返回一个空的map实例
func NotNullMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return make(map[K]V)
	}
	return m
}

// NotNullMapPtr 检查指向map的指针是否为nil，如果是，则返回一个指向空map实例的指针
func NotNullMapPtr[K comparable, V any](m *map[K]V) map[K]V {
	if m == nil || *m == nil {
		return make(map[K]V)
	}
	return *m
}

func Canceled(ctx context.Context) bool {
	if ctx.Done() == nil {
		return false
	}
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
