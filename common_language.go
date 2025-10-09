package commonlanguage

import (
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/spf13/cast"
	"github.com/suifengpiao14/sqlbuilder"
)

// NewId 更新时必填(非必填场景可以不引入改字段)更新只会出现在where中，不会出现在set 中, 查询时可选，支持,分割多个
func NewId[T int | uint | int64 | uint64 | []int | []int64 | []uint64](autoId T) (field *sqlbuilder.Field) {
	field = sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return autoId, nil })
	field.SetName("id").SetTitle("ID").MergeSchema(sqlbuilder.Schema{
		Type:      sqlbuilder.Schema_Type_int,
		Maximum:   sqlbuilder.Int_maximum_bigint,
		MaxLength: 64,
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

func NewStringId(id string) (field *sqlbuilder.Field) {
	field = sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return id, nil })
	field.SetName("id").SetTitle("ID").MergeSchema(sqlbuilder.Schema{
		Type:      sqlbuilder.Schema_Type_string,
		MaxLength: 64,
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
	f = sqlbuilder.NewIntField(pageIndex, "pageIndex", "页码", 0).SetTag(sqlbuilder.Field_tag_pageIndex).Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnShieldForData)
		f.WhereFns.Reset()
	})
	return f
}

func NewPageSize(pageSize int) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewField(pageSize).SetName("pageSize").SetTitle("每页数量").SetTag(sqlbuilder.Field_tag_pageSize).Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		// 有可能用于update 时 的limit 值，所以需要屏蔽更新以及where条件
		f.ValueFns.Append(sqlbuilder.ValueFnShieldForData)
		f.WhereFns.Reset()
	})
	return f
}

func NewUpdateLimit(updateLimit int) *sqlbuilder.Field {
	return sqlbuilder.NewIntField(updateLimit, "updateLimit", "更新语句的limit", 0).SetTag(sqlbuilder.Field_tag_pageSize).Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnDBFormat(sqlbuilder.ValueFnShieldFn))
	})
}

func NewDateTime(dateTime string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(dateTime, "dateTime", "日期时间", 20)
}

func NewCreatedAt(time string) (f *sqlbuilder.Field) {
	f = NewTime(time).SetName("createdAt").SetTitle("创建时间").SetTag(sqlbuilder.Tag_createdAt)
	f.SceneUpdate(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnShield) // 更新时屏蔽
	})
	return f
}
func NewCompletedAt(time string) (f *sqlbuilder.Field) {
	f = NewTime(time).SetName("completedAt").SetTitle("完成时间")
	f.Apply(sqlbuilder.ApplyFnValueFnSetIfEmpty) // 为空才更新,使用第一次数据
	return f
}
func NewUpdatedAt(time string) (f *sqlbuilder.Field) {
	f = NewTime(time).SetName("updatedAt").SetTitle("更新时间").SetTag(sqlbuilder.Tag_updatedAt)
	return f
}

func makeUpkdateLock() (updateLock string) {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func NewUpdateLock(updateLock string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(updateLock, "updateLock", "乐观更新锁", 64).Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		// 写入数据库时，使用及时生成
		f.ValueFns.Append(sqlbuilder.ValueFnOnlyForData(func(inputValue any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) {
			return makeUpkdateLock(), nil
		}))
		f.SceneUpdate(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil) // 传入的值，只作用于查询条件，不存在跳过该条件 ,并且限定在update操作场景中
		})
	})
}

func NewStatusWithDeleted[T int | string](status T, deletedStatus T, enums ...sqlbuilder.Enum) *sqlbuilder.Field {
	return sqlbuilder.NewField(status).SetName("status").SetTitle("状态").AppendEnum(enums...).Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil) // 查询的时候,可以为空,需要在验证前格式化数据，为空直接设置nil
		})

		deletedWhereValueFnForSelect := sqlbuilder.ValueFn{ // 查询的时候,过滤删除的列
			Layer: sqlbuilder.Value_Layer_OnlyForData,
			Fn: func(inputValue any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) {
				expression := goqu.I(f.DBColumnName().FullName()).Neq(deletedStatus)
				return expression, nil
			},
		}

		f.SceneInsert(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			//f.ValueFns.Append(sqlbuilder.ValueFnShieldForData) //新增时,忽略该字段
			//2025-06-26 11:03  新增不能屏蔽，应该从Table.Columns 里面筛选默认值，而不是屏蔽
			f.ValueFns.ResetSetValueFn(func(inputValue any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) {
				//value := f.GetTable().Columns.GetByFieldNameMust(f.Name).Default
				var value any
				c, ok := f.GetTable().Columns.GetByFieldName(f.Name) // 这里为了兼容 历史没设置table.AddColumn 情况，不使用GetByFieldNameMust
				if ok {
					value = c.Default
				}
				return value, nil
			})
		})

		f.SceneUpdate(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			f.ValueFns.Append(sqlbuilder.ValueFnShieldForData) //更新时,忽略该字段,如果更新时确实需要更新该字段，则另外NewStatus 处理，NewStatusWithDeleted偏重于考虑软删除
		})

		f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			f.WhereFns.Append(deletedWhereValueFnForSelect)
		})

		//设置删除场景
		f.SceneFn(sqlbuilder.SceneFn{
			Scene: sqlbuilder.SCENE_SQL_DELETE,
			Fn: func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
				f.ValueFns.ResetSetValueFn(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) {
					return deletedStatus, nil
				})
				f.WhereFns.Append(deletedWhereValueFnForSelect)
			},
		})

	}).SetFieldName(sqlbuilder.Field_name_deletedAt) // 标记为删除字段
}

// NewDeletedAt 通过删除时间列标记删除
func NewDeletedAt() (f *sqlbuilder.Field) {
	return NewDeletedWithValue(func() any { return "" }, func() any { return time.Now().Local().Format(time.DateTime) })
}

const (
	Deleted_effect_value_null = "null"
)

// NewDeletedWithValue 通过删除时间列标记删除,增加默认值参数，方便兼容一些数据库的默认值为0000-00-00 00:00:00的情况
func NewDeletedWithValue(okValueFn func() any, deletedValueFn func() any) (f *sqlbuilder.Field) { // 有的删除列默认值使用 0000-00-00 00:00:00 作为有效值，所以增加这个方法
	//对于status 字段 0表示删除 1,2,3... 表示有效则okValueFn=func(){return goqu.I(f.DBColumnName().FullName()).Neq(0)}
	whereValueFormatFn := func(inputValue any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) {
		deletedColumnName := f.Name
		deletedFields := sqlbuilder.Fields(fs).Fielter(func(f sqlbuilder.Field) bool {
			return f.Name == deletedColumnName
		})
		if len(deletedFields) > 1 {
			return nil, nil // 理论上不应该出现这种情况，如果出现，则忽略该字段的过滤条件(场景场景:status字段也是删除字段)外部会传入status条件,这里就不增加删除条件
		}
		if _, ok := inputValue.(goqu.Expression); ok { // 兼容传入表达式的情况，比如goqu.C("deleted_at").IsNull() ，由于这个函数再f.SceneSelect 内添加，所以会最后运行
			return inputValue, nil
		}
		var value = okValueFn()
		if cast.ToString(value) == Deleted_effect_value_null {
			value = goqu.I(f.DBColumnName().FullName()).IsNull()
		}
		return value, nil
	}
	f = NewTime("").SetName("deleted_at").SetTitle("删除时间").SetFieldName(sqlbuilder.Field_name_deletedAt) // 标记为删除字段

	f.SceneInsert(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnShieldForData)
	})
	f.SceneUpdate(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnShieldForData)
	})
	f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.WhereFns.Append(sqlbuilder.ValueFnDBFormat(whereValueFormatFn))
	})

	//设置删除场景
	f.SceneFn(sqlbuilder.SceneFn{
		Scene: sqlbuilder.SCENE_SQL_DELETE,
		Fn: func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			f.ValueFns.ResetSetValueFn(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) {
				return deletedValueFn(), nil
			})
			f.WhereFns.ResetSetValueFn(whereValueFormatFn)
		},
	})
	return f
}

// deprecated use   NewDeletedWithValue 通过删除时间列标记删除,增加默认值参数，方便兼容一些数据库的默认值为0000-00-00 00:00:00的情况
func NewDeletedWithEffectValue(effectValue string) (f *sqlbuilder.Field) {
	return NewDeletedWithValue(func() any { return effectValue }, func() any { return time.Now().Local().Format(time.DateTime) })
}

func NewCreateTime(createTime string) *sqlbuilder.Field {
	return NewCreatedAt(createTime).SetName("createTime")
}

func NewUpdateTime(updateTime string) *sqlbuilder.Field {
	return NewUpdatedAt(updateTime).SetName("updateTime")
}

func NewFileName(fileName string) *sqlbuilder.Field {
	return sqlbuilder.NewStringField(fileName, "fileName", "文件名", 256) // 文件名支持url长度
}

func NewStatus[T int | string](status T, enums sqlbuilder.Enums) *sqlbuilder.Field {
	return sqlbuilder.NewField(status).SetName("status").SetTitle("状态").AppendEnum(enums...).Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil) // 查询的时候,可以为空,需要在验证前格式化数据，为空直接设置nil
		})
	})
}

func NewTime(time string) *sqlbuilder.Field {
	f := sqlbuilder.NewField(time).SetName("time").SetTitle("时间").SetType(sqlbuilder.Schema_Type_string).SetFormat(sqlbuilder.Schema_format_dateTime).MergeSchema(sqlbuilder.Schema{
		MaxLength: 20, // 2006-01-02 15:04:05 19个字符
	})
	return f
}

func NewCreatedAtBegin(time string) *sqlbuilder.Field {
	f := NewTime(time).SetName("createdAtBegin").SetTitle("创建时间开始点")
	f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	f.Apply(sqlbuilder.ApplyFnWhereGte)
	return f
}

func NewCreateAtEnd(time string) *sqlbuilder.Field {
	f := NewTime(time).SetName("createdAtEnd").SetTitle("创建时间结束点")
	f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	f.Apply(sqlbuilder.ApplyFnWhereLte)
	return f
}
func NewUpdatedAtBegin(time string) *sqlbuilder.Field {
	f := NewTime(time).SetName("updatedAtBegin").SetTitle("更新时间开始点")
	f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	f.Apply(sqlbuilder.ApplyFnWhereGte)
	return f
}

func NewUpdatedAtEnd(time string) *sqlbuilder.Field {
	f := NewTime(time).SetName("updatedAtEnd").SetTitle("更新时间结束点")
	f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	f.Apply(sqlbuilder.ApplyFnWhereLte)
	return f
}
func NewCreateTimeBegin(time string) *sqlbuilder.Field {
	f := NewCreateAtEnd(time).SetName("createdTimeBegin")
	return f
}

func NewCreateTimeEnd(time string) *sqlbuilder.Field {
	f := NewCreateAtEnd(time).SetName("createdTimeEnd")
	return f
}

const (
	Schema_format_email = "email"
	Schema_format_phone = "phone"
)

func NewEmail(email string) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return email, nil }).SetName("email").SetTitle("邮箱").SetFormat(Schema_format_email)
	f.MergeSchema(sqlbuilder.Schema{
		Type:      sqlbuilder.Schema_Type_string,
		MaxLength: 32,
		RegExp:    `([A-Za-z0-9\-]+\.)+[A-Za-z]{2,6}`, // 邮箱验证表达式
	})
	f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil) // 由于value 的 validate 在 whereFn 之前，所以这里需要设置ValueFns
	})
	return f
}

func NewPhone(phone string) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return phone, nil })
	f.SetName("phone").SetTitle("手机号").MergeSchema(sqlbuilder.Schema{
		Type:      sqlbuilder.Schema_Type_string,
		MaxLength: 15,
		RegExp:    `^1[3-9]\d{9}$`, // 中国大陆手机号正则表达式
		Format:    Schema_format_phone,
	})
	f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil) // 由于value 的 validate 在 whereFn 之前，所以这里需要设置ValueFns
	})

	return f
}

type EnumField struct {
	Enums sqlbuilder.Enums
	Field *sqlbuilder.Field
}

func NewEnumField(value any, enums sqlbuilder.Enums) *EnumField {
	e := &EnumField{
		Enums: enums,
	}
	e.Field = sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return value, nil }).SetName("enum_column").SetTag("枚举列")
	e.Field.AppendEnum(enums...)
	return e
}

func NewGender[T int | string](val T, man T, woman T) *EnumField {
	genderField := NewEnumField(val, sqlbuilder.Enums{
		sqlbuilder.Enum{
			Key:   man,
			Title: "男",
		},
		sqlbuilder.Enum{
			Key:   woman,
			Title: "女",
		},
	})
	genderField.Field.SetName("gender").SetTitle("性别").Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil)
		f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			f.Schema.Enums.Append(sqlbuilder.Enum{
				Key:   "",
				Title: "全部",
			})
			f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil)
		})
	})
	return genderField
}

func NewBooleanField[T int | string](val T, enumTrue T, enumFalse T) *EnumField {
	genderField := NewEnumField(val, sqlbuilder.Enums{
		sqlbuilder.Enum{
			Key:   enumTrue,
			Title: "真",
		},
		sqlbuilder.Enum{
			Key:   enumFalse,
			Title: "假",
		},
	})
	genderField.Field.SetName("bool").SetTitle("真假").Apply(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil)
		f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
			f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil)
		})
	})
	return genderField
}

func NewAddress(address string) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return address, nil }).SetName("address").SetTitle("地址").MergeSchema(sqlbuilder.Schema{
		Type:      sqlbuilder.Schema_Type_string,
		MaxLength: 128, // 线上统计最大55个字符，设置128 应该适合大部分场景大小
	})
	f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	return f
}

func NewHeight(height int) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return height, nil }).SetName("height").SetTitle("高").MergeSchema(sqlbuilder.Schema{
		Type:      sqlbuilder.Schema_Type_int,
		MaxLength: 10000, //日常物体、人、动物高不操过1万m/cm
	})
	f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	return f
}

func NewOwnerID[T int | string | int64 | []int | []string | []int64](value T) *sqlbuilder.Field {
	field := sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return value, nil }).SetName("ownerId").SetTitle("所有者").MergeSchema(sqlbuilder.Schema{
		Comment:      "所有者ID",
		Type:         sqlbuilder.Schema_Type_string,
		MaxLength:    64,
		ShieldUpdate: true, // 所有者不可跟新
	})
	field.SceneInsert(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.SetRequired(true)
	})
	field.ShieldUpdate(true) // 所有者不能更换,当前记录是所有者属性的描述，所有者发生变更，本条记录失去业务意义
	return field
}

func NewUserId[T int | string | int64 | []int | []string | []int64](userId T) *sqlbuilder.Field {
	f := sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return userId, nil }).SetName("userId").SetTitle("用户ID")
	return f
}

func NewIdentifier(value any) *sqlbuilder.Field {
	f := sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return value, nil }).SetName("identity").SetTitle("标识")
	return f
}

func NewTitle(value string) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewField(func(in any, f *sqlbuilder.Field, fs ...*sqlbuilder.Field) (any, error) { return value, nil })
	f.SetName("title").SetTitle("标题").MergeSchema(sqlbuilder.Schema{
		Type:      sqlbuilder.Schema_Type_string,
		MaxLength: 64,
	}).ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil)

	f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.WhereFns.Append(sqlbuilder.ValueFnWhereLike)
	})
	return f
}
func NewTag(tags string) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewStringField(tags, "tag", "标签", 128)
	f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil)
		f.Apply(sqlbuilder.ApplyFnWhereFindInColumnSet) // 标签支持结合查询
	})
	return f
}

func NewClassify(classfiy string) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewStringField(classfiy, "classify", "分类", 64)
	f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	return f
}
func NewFullData(fullData string) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewStringField(fullData, "fullData", "全量数据", 0).Comment("使用json格式存储记录所有数据(自增id除外)主要用于后续数据异构")
	f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil)
	f.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.WhereFns.Append(sqlbuilder.ValueFnShield) // 全量数据为json格式，不支持查询
	})
	return f
}

func NewRemark(remark string) (f *sqlbuilder.Field) {
	f = sqlbuilder.NewStringField(remark, "remark", "备注", 0)
	return f
}

func NewOperatorName(operatorName string) *sqlbuilder.Field {
	f := sqlbuilder.NewStringField(operatorName, "operatorName", "操作人", 32)
	f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil) // 操作人为空，不更新
	return f
}

func NewOperatorId[T int | int64 | string](operatorId T) (f *sqlbuilder.Field) {
	a := any(operatorId)
	switch v := a.(type) {
	case int, int64:
		f = sqlbuilder.NewField(operatorId).SetName("operatorId").SetTitle("操作人ID").MergeSchema(sqlbuilder.Schema{Maximum: sqlbuilder.UnsinedInt_maximum_bigint})
	case string:
		f = sqlbuilder.NewStringField(v, "operatorId", "操作人", 64) // 字符串类型，需要设置最大长度

	}
	f.ValueFns.Append(sqlbuilder.ValueFnEmpty2Nil) // 操作人为空，不更新
	return f
}
