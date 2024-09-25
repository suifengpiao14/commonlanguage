package commonlanguage

import "github.com/suifengpiao14/sqlbuilder"

//NewAutoId 数据库自增ID insert 时，sql中不设置值，更新时必填(非必填场景可以不引入改字段)更新只会出现在where中，不会出现在set 中, 查询时可选，支持,分割多个
func NewAutoId[T int | int64 | uint64 | []int | []int64 | []uint64](autoId uint) (field *sqlbuilder.Field) {
	field = sqlbuilder.NewField(func(in any) (any, error) { return autoId, nil })
	field.SetName("id").SetTitle("ID").MergeSchema(sqlbuilder.Schema{
		Type:          sqlbuilder.Schema_Type_int,
		Maximum:       sqlbuilder.Int_maximum_bigint,
		MaxLength:     64,
		Primary:       true,
		AutoIncrement: true,
	})

	field.SceneInsert(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ValueFns.Append(sqlbuilder.ValueFnShield)
	})
	field.SceneUpdate(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.ShieldUpdate(true).SetRequired(true) // id 不能更新
		f.WhereFns.Append(sqlbuilder.ValueFnFormatArray)
		f.MergeSchema(sqlbuilder.Schema{
			Minimum: 1,
		})
	})

	field.SceneSelect(func(f *sqlbuilder.Field, fs ...*sqlbuilder.Field) {
		f.WhereFns.Append(sqlbuilder.ValueFnEmpty2Nil, sqlbuilder.ValueFnFormatArray)
		if f.Schema.Required {
			f.MergeSchema(sqlbuilder.Schema{
				Minimum: 1,
			})
		}
	})
	return field
}
