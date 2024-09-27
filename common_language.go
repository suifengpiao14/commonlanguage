package commonlanguage

import (
	"time"

	"github.com/suifengpiao14/sqlbuilder"
)

// NewId 更新时必填(非必填场景可以不引入改字段)更新只会出现在where中，不会出现在set 中, 查询时可选，支持,分割多个
func NewId[T int | int64 | uint64 | []int | []int64 | []uint64](autoId T) (field *sqlbuilder.Field) {
	field = sqlbuilder.NewField(func(in any) (any, error) { return autoId, nil })
	field.SetName("id").SetTitle("ID").MergeSchema(sqlbuilder.Schema{
		Type:          sqlbuilder.Schema_Type_int,
		Maximum:       sqlbuilder.Int_maximum_bigint,
		MaxLength:     64,
		Primary:       true,
		AutoIncrement: true,
	})
	field.SceneUpdate(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ShieldUpdate(true).SetRequired(true) // id 不能更新
		f.WhereFns.Append(sqlbuilder.ValueFnForward, sqlbuilder.ValueFnFormatArray)

	})
	field.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil, sqlbuilder.ValueFnFormatArray)
	})
	return field
}

func NewPageIndex(pageIndex int) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewIntField(pageIndex, "pageIndex", "页码", 0).SetTag(sqlbuilder.Field_tag_pageIndex)
	return f
}

func NewPageSize(pageSize int) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewField(pageSize).SetName("pageSize").SetTitle("每页数量").SetTag(sqlbuilder.Field_tag_pageSize)
	return f
}

func NewDateTime(dateTime string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(dateTime, "dateTime", "日期时间", 20)
}

func NewCreatedAt() (f *sqlbuilder.Field) {
	f = sqlbuilder.NewField(func(in any) (any, error) {
		return time.Now().Local().Format(time.DateTime), nil
	}).SetName("created_at").SetTitle("创建时间").SetTag(sqlbuilder.Tag_createdAt)
	f.SceneUpdate(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnShield) // 更新时屏蔽
	})
	return f
}
func NewUpdatedAt() (f *sqlbuilder.Field) {
	f = sqlbuilder.NewField(func(in any) (any, error) {
		return time.Now().Local().Format(time.DateTime), nil
	})
	f.SetName("updated_at").SetTitle("更新时间").SetTag(sqlbuilder.Tag_updatedAt)
	return f
}

func NewCreateTime(createTime string) *sqlbuilder.Field {
	return NewCreatedAt().SetName("createTime")
}

func NewUpdateTime(updateTime string) *sqlbuilder.Field {
	return NewUpdatedAt().SetName("updateTime")
}

func NewFileName(fileName string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(fileName, "fileName", "文件名", 256) // 文件名支持url长度
}

func NewStatus[T int | string](status T, enums sqlbuilder.Enums) *sqlbuilder.Field {
	return sqlbuilder.NewField(status).SetName("status").SetTitle("状态").AppendEnum(enums...).Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil) // 查询的时候,可以为空,需要再验证前格式化数据，为空直接设置nil
		})
	})
}

func NewCreatedAtBegin(createdAtBegin string, dbFieldName string) *sqlbuilder.Field {
	f := sqlbuilder.NewField(createdAtBegin).SetName("createdAtBegin").SetTitle("创建时间开始点").SetDBName(dbFieldName).SetType(sqlbuilder.Schema_Type_string)
	f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	f.Apply(sqlbuilder.ApplyFnWhereGte)
	return f
}

func NewCreateAtEnd(createdAtEnd string, dbFieldName string) *sqlbuilder.Field {
	f := sqlbuilder.NewField(createdAtEnd).SetName("createdAtEnd").SetTitle("创建时间结束点").SetDBName(dbFieldName).SetType(sqlbuilder.Schema_Type_string)
	f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	f.Apply(sqlbuilder.ApplyFnWhereLte)
	return f
}
func NewCreateTimeBegin(createdTimeBegin string, dbFieldName string) *sqlbuilder.Field {
	f := NewCreateAtEnd(createdTimeBegin, dbFieldName).SetName("createdTimeBegin")
	return f
}

func NewCreateTimeEnd(createdTimeEnd string, dbFieldName string) *sqlbuilder.Field {
	f := NewCreateAtEnd(createdTimeEnd, dbFieldName).SetName("createdTimeEnd")
	return f
}
