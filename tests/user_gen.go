package tests

import (
	"database/sql"

	"github.com/monochromegane/argen"
)

var db *sql.DB

func Use(DB *sql.DB) {
	db = DB
}

func (m *User) fieldByName(name string) interface{} {
	switch name {
	case "id":
		return &m.Id
	case "name":
		return &m.Name
	default:
		return ""
	}
}

func (m *User) fieldsByName(names []string) []interface{} {
	fields := []interface{}{}
	for _, n := range names {
		f := m.fieldByName(n)
		fields = append(fields, f)
	}
	return fields
}

func (m User) Select(columns ...string) *UserRelation {
	r := m.newRelation()
	r.Relation.Columns(columns...)
	return r
}

func (m User) Find(id int) (*User, error) {
	return m.newRelation().Find(id)
}

func (r *UserRelation) Find(id int) (*User, error) {
	return r.Where("id", id).QueryRow()
}

type UserParams User

func (m User) Create(p UserParams) (*User, *ar.Errors) {
	n := &User{
		Id:   p.Id,
		Name: p.Name,
	}
	_, errs := n.Save()
	return n, errs
}

func (m *User) IsNewRecord() bool {
	return ar.IsZero(m.Id)
}

func (m *User) IsPersistent() bool {
	return !m.IsNewRecord()
}

func (m *User) Save() (bool, *ar.Errors) {
	if ok, errs := m.IsValid(); !ok {
		return false, errs
	}
	errs := &ar.Errors{}
	if m.IsNewRecord() {
		ins := ar.NewInsert()
		q, b := ins.Table("users").Params(map[string]interface{}{
			"name": m.Name,
		}).Build()

		if _, err := db.Exec(q, b...); err != nil {
			errs.Add("base", err)
			return false, errs
		}
		return true, nil
	} else {
		upd := ar.NewUpdate()
		q, b := upd.Table("users").Params(map[string]interface{}{
			"id":   m.Id,
			"name": m.Name,
		}).Where("id", m.Id).Build()

		if _, err := db.Exec(q, b...); err != nil {
			errs.Add("base", err)
			return false, errs
		}
		return true, nil
	}
}

type UserRelation struct {
	src *User
	*ar.Relation
}

func (m *User) newRelation() *UserRelation {
	r := ar.NewRelation()
	r.Table("users").Columns(
		"id",
		"name",
	)

	return &UserRelation{m, r}
}

func (r *UserRelation) Query() ([]*User, error) {
	q, b := r.Build()
	rows, err := db.Query(q, b...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*User{}
	for rows.Next() {
		row := &User{}
		err := rows.Scan(row.fieldsByName(r.Relation.GetColumns())...)
		if err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, nil
}

func (r *UserRelation) QueryRow() (*User, error) {
	q, b := r.Build()
	row := &User{}
	err := db.QueryRow(q, b...).Scan(row.fieldsByName(r.Relation.GetColumns())...)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (m User) First() (*User, error) {
	return m.newRelation().First()
}

func (r *UserRelation) First() (*User, error) {
	return r.Order("id", "ASC").Limit(1).QueryRow()
}

func (m User) Last() (*User, error) {
	return m.newRelation().Last()
}

func (r *UserRelation) Last() (*User, error) {
	return r.Order("id", "DESC").Limit(1).QueryRow()
}

func (m User) Where(cond string, args ...interface{}) *UserRelation {
	return m.newRelation().Where(cond, args...)
}

func (r *UserRelation) Where(cond string, args ...interface{}) *UserRelation {
	r.Relation.Where(cond, args...)
	return r
}

func (r *UserRelation) And(cond string, args ...interface{}) *UserRelation {
	r.Relation.And(cond, args...)
	return r
}

func (r *UserRelation) Order(column, order string) *UserRelation {
	r.Relation.OrderBy(column, order)
	return r
}

func (r *UserRelation) Limit(limit int) *UserRelation {
	r.Relation.Limit(limit)
	return r
}

func (r *UserRelation) Offset(offset int) *UserRelation {
	r.Relation.Offset(offset)
	return r
}

func (r *UserRelation) Group(group string, groups ...string) *UserRelation {
	r.Relation.GroupBy(group, groups...)
	return r
}

func (r *UserRelation) Having(cond string, args ...interface{}) *UserRelation {
	r.Relation.Having(cond, args...)
	return r
}

func (r *UserRelation) Explain() *UserRelation {
	r.Relation.Explain()
	return r
}

func (m User) DeleteAll() (bool, *ar.Errors) {
	errs := &ar.Errors{}
	del := ar.NewDelete()
	del.Table("users")
	q, b := del.Build()
	if _, err := db.Exec(q, b...); err != nil {
		errs.Add("base", err)
		return false, errs
	}
	return true, nil
}

func (m User) IsValid() (bool, *ar.Errors) {
	result := true
	errors := &ar.Errors{}
	rules := map[string]*ar.Validation{}
	for name, rule := range rules {
		if ok, errs := ar.NewValidator(rule).IsValid(m.fieldByName(name)); !ok {
			result = false
			errors.Set(name, errs)
		}
	}
	customs := []ar.CustomValidator{}
	for _, c := range customs {
		if ok, column, err := c(); !ok {
			result = false
			errors.Add(column, err)
		}
	}
	return result, errors
}