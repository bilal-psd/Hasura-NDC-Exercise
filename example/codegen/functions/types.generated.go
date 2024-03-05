// Code generated by github.com/hasura/ndc-sdk-go/codegen, DO NOT EDIT.
package functions

import (
	"github.com/google/uuid"
	"github.com/hasura/ndc-sdk-go/utils"
)

// FromValue decodes values from map
func (j *GetTypesArguments) FromValue(input map[string]any) error {
	var err error
	err = utils.DecodeObjectValue(&j.ArrayObject, input, "ArrayObject")
	if err != nil {
		return err
	}
	j.ArrayObjectPtr = new([]struct {
		Content string "json:\"content\""
	})
	err = utils.DecodeNullableObjectValue(j.ArrayObjectPtr, input, "ArrayObjectPtr")
	if err != nil {
		return err
	}
	j.Bool, err = utils.GetBool(input, "Bool")
	if err != nil {
		return err
	}
	j.BoolPtr, err = utils.GetNullableBool(input, "BoolPtr")
	if err != nil {
		return err
	}
	err = utils.DecodeObjectValue(&j.CustomScalar, input, "CustomScalar")
	if err != nil {
		return err
	}
	j.CustomScalarPtr = new(CommentText)
	err = utils.DecodeNullableObjectValue(j.CustomScalarPtr, input, "CustomScalarPtr")
	if err != nil {
		return err
	}
	j.Duration, err = utils.GetDuration(input, "Duration")
	if err != nil {
		return err
	}
	j.DurationPtr, err = utils.GetNullableDuration(input, "DurationPtr")
	if err != nil {
		return err
	}
	j.Float32, err = utils.GetFloat[float32](input, "Float32")
	if err != nil {
		return err
	}
	j.Float32Ptr, err = utils.GetNullableFloat[float32](input, "Float32Ptr")
	if err != nil {
		return err
	}
	j.Float64, err = utils.GetFloat[float64](input, "Float64")
	if err != nil {
		return err
	}
	j.Float64Ptr, err = utils.GetNullableFloat[float64](input, "Float64Ptr")
	if err != nil {
		return err
	}
	j.Int, err = utils.GetInt[int](input, "Int")
	if err != nil {
		return err
	}
	j.Int16, err = utils.GetInt[int16](input, "Int16")
	if err != nil {
		return err
	}
	j.Int16Ptr, err = utils.GetNullableInt[int16](input, "Int16Ptr")
	if err != nil {
		return err
	}
	j.Int32, err = utils.GetInt[int32](input, "Int32")
	if err != nil {
		return err
	}
	j.Int32Ptr, err = utils.GetNullableInt[int32](input, "Int32Ptr")
	if err != nil {
		return err
	}
	j.Int64, err = utils.GetInt[int64](input, "Int64")
	if err != nil {
		return err
	}
	j.Int64Ptr, err = utils.GetNullableInt[int64](input, "Int64Ptr")
	if err != nil {
		return err
	}
	j.Int8, err = utils.GetInt[int8](input, "Int8")
	if err != nil {
		return err
	}
	j.Int8Ptr, err = utils.GetNullableInt[int8](input, "Int8Ptr")
	if err != nil {
		return err
	}
	j.IntPtr, err = utils.GetNullableInt[int](input, "IntPtr")
	if err != nil {
		return err
	}
	err = utils.DecodeObjectValue(&j.NamedArray, input, "NamedArray")
	if err != nil {
		return err
	}
	j.NamedArrayPtr = new([]Author)
	err = utils.DecodeNullableObjectValue(j.NamedArrayPtr, input, "NamedArrayPtr")
	if err != nil {
		return err
	}
	err = utils.DecodeObjectValue(&j.NamedObject, input, "NamedObject")
	if err != nil {
		return err
	}
	j.NamedObjectPtr = new(Author)
	err = utils.DecodeNullableObjectValue(j.NamedObjectPtr, input, "NamedObjectPtr")
	if err != nil {
		return err
	}
	err = utils.DecodeObjectValue(&j.Object, input, "Object")
	if err != nil {
		return err
	}
	j.ObjectPtr = new(struct {
		Long int
		Lat  int
	})
	err = utils.DecodeNullableObjectValue(j.ObjectPtr, input, "ObjectPtr")
	if err != nil {
		return err
	}
	j.String, err = utils.GetString(input, "String")
	if err != nil {
		return err
	}
	j.StringPtr, err = utils.GetNullableString(input, "StringPtr")
	if err != nil {
		return err
	}
	j.Time, err = utils.GetDateTime(input, "Time")
	if err != nil {
		return err
	}
	j.TimePtr, err = utils.GetNullableDateTime(input, "TimePtr")
	if err != nil {
		return err
	}
	err = utils.DecodeObjectValue(&j.UUID, input, "UUID")
	if err != nil {
		return err
	}
	j.UUIDPtr = new(uuid.UUID)
	err = utils.DecodeNullableObjectValue(j.UUIDPtr, input, "UUIDPtr")
	if err != nil {
		return err
	}
	j.Uint, err = utils.GetUint[uint](input, "Uint")
	if err != nil {
		return err
	}
	j.Uint16, err = utils.GetUint[uint16](input, "Uint16")
	if err != nil {
		return err
	}
	j.Uint16Ptr, err = utils.GetNullableUint[uint16](input, "Uint16Ptr")
	if err != nil {
		return err
	}
	j.Uint32, err = utils.GetUint[uint32](input, "Uint32")
	if err != nil {
		return err
	}
	j.Uint32Ptr, err = utils.GetNullableUint[uint32](input, "Uint32Ptr")
	if err != nil {
		return err
	}
	j.Uint64, err = utils.GetUint[uint64](input, "Uint64")
	if err != nil {
		return err
	}
	j.Uint64Ptr, err = utils.GetNullableUint[uint64](input, "Uint64Ptr")
	if err != nil {
		return err
	}
	j.Uint8, err = utils.GetUint[uint8](input, "Uint8")
	if err != nil {
		return err
	}
	j.Uint8Ptr, err = utils.GetNullableUint[uint8](input, "Uint8Ptr")
	if err != nil {
		return err
	}
	j.UintPtr, err = utils.GetNullableUint[uint](input, "UintPtr")
	if err != nil {
		return err
	}
	return nil
}

// FromValue decodes values from map
func (j *GetArticlesArguments) FromValue(input map[string]any) error {
	var err error
	j.Limit, err = utils.GetFloat[float64](input, "Limit")
	if err != nil {
		return err
	}
	return nil
}

// ToMap encodes the struct to a value map
func (j Author) ToMap() map[string]any {
	result := map[string]any{
		"created_at": j.CreatedAt,
		"duration":   j.Duration,
		"id":         j.ID,
	}
	return result
}

// ToMap encodes the struct to a value map
func (j CreateArticleResult) ToMap() map[string]any {
	result := map[string]any{
		"authors": utils.EncodeMaps(j.Authors),
		"id":      j.ID,
	}
	return result
}

// ToMap encodes the struct to a value map
func (j CreateAuthorResult) ToMap() map[string]any {
	result := map[string]any{
		"created_at": j.CreatedAt,
		"id":         j.ID,
		"name":       j.Name,
	}
	return result
}

// ToMap encodes the struct to a value map
func (j GetArticlesResult) ToMap() map[string]any {
	result := map[string]any{
		"id":   j.ID,
		"Name": j.Name,
	}
	return result
}

// ToMap encodes the struct to a value map
func (j GetTypesArguments) ToMap() map[string]any {

	var result_ObjectPtr map[string]any
	if j.ObjectPtr != nil {
		result_ObjectPtr = map[string]any{
			"Lat":  j.ObjectPtr.Lat,
			"Long": j.ObjectPtr.Long,
		}
	}
	result_Object := map[string]any{
		"created_at": j.Object.CreatedAt,
		"id":         j.Object.ID,
	}
	var result_ArrayObjectPtr []map[string]any
	if j.ArrayObjectPtr != nil {
		result_ArrayObjectPtr = make([]map[string]any, len(*j.ArrayObjectPtr))
		for i, _item := range *j.ArrayObjectPtr {
			item := map[string]any{
				"content": _item.Content,
			}
			result_ArrayObjectPtr[i] = item
		}
	}
	var result_ArrayObject []map[string]any
	result_ArrayObject = make([]map[string]any, len(j.ArrayObject))
	for i, _item := range j.ArrayObject {
		item := map[string]any{
			"content": _item.Content,
		}
		result_ArrayObject[i] = item
	}
	result := map[string]any{
		"ArrayObject":     result_ArrayObject,
		"ArrayObjectPtr":  result_ArrayObjectPtr,
		"Bool":            j.Bool,
		"BoolPtr":         j.BoolPtr,
		"CustomScalar":    j.CustomScalar,
		"CustomScalarPtr": j.CustomScalarPtr,
		"Duration":        j.Duration,
		"DurationPtr":     j.DurationPtr,
		"Float32":         j.Float32,
		"Float32Ptr":      j.Float32Ptr,
		"Float64":         j.Float64,
		"Float64Ptr":      j.Float64Ptr,
		"Int":             j.Int,
		"Int16":           j.Int16,
		"Int16Ptr":        j.Int16Ptr,
		"Int32":           j.Int32,
		"Int32Ptr":        j.Int32Ptr,
		"Int64":           j.Int64,
		"Int64Ptr":        j.Int64Ptr,
		"Int8":            j.Int8,
		"Int8Ptr":         j.Int8Ptr,
		"IntPtr":          j.IntPtr,
		"NamedArray":      utils.EncodeMaps(j.NamedArray),
		"NamedArrayPtr":   utils.EncodeNullableMaps(j.NamedArrayPtr),
		"NamedObject":     utils.EncodeMap(j.NamedObject),
		"NamedObjectPtr":  utils.EncodeMap(j.NamedObjectPtr),
		"Object":          result_Object,
		"ObjectPtr":       result_ObjectPtr,
		"String":          j.String,
		"StringPtr":       j.StringPtr,
		"Time":            j.Time,
		"TimePtr":         j.TimePtr,
		"UUID":            j.UUID,
		"UUIDPtr":         j.UUIDPtr,
		"Uint":            j.Uint,
		"Uint16":          j.Uint16,
		"Uint16Ptr":       j.Uint16Ptr,
		"Uint32":          j.Uint32,
		"Uint32Ptr":       j.Uint32Ptr,
		"Uint64":          j.Uint64,
		"Uint64Ptr":       j.Uint64Ptr,
		"Uint8":           j.Uint8,
		"Uint8Ptr":        j.Uint8Ptr,
		"UintPtr":         j.UintPtr,
	}
	return result
}

// ToMap encodes the struct to a value map
func (j HelloResult) ToMap() map[string]any {
	result := map[string]any{
		"foo":  j.Foo,
		"id":   j.ID,
		"num":  j.Num,
		"text": j.Text,
	}
	return result
}

// ScalarName get the schema name of the scalar
func (j CommentText) ScalarName() string {
	return "CommentString"
}

// ScalarName get the schema name of the scalar
func (j ScalarFoo) ScalarName() string {
	return "Foo"
}
