package tests

import (
	"database/sql"
	"log"
	"os"
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/monochromegane/argen"
)

func TestMain(m *testing.M) {
	db, err := testDb()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	Use(db)
	LogMode(true)
	for _, q := range testTables() {
		_, err = db.Exec(q)
		if err != nil {
			log.Fatal(err, q)
		}
	}

	os.Exit(m.Run())
}

func TestSelect(t *testing.T) {
	u := &User{Name: "test"}
	u.Save()
	defer User{}.DeleteAll()

	u, err := User{}.Select("id").First()
	assertError(t, err)

	if !ar.IsZero(u.Name) {
		t.Errorf("column value should be empty, but %s", u.Name)
	}
}

func TestFind(t *testing.T) {
	expect := &User{Name: "test"}
	expect.Save()
	defer User{}.DeleteAll()

	u, err := User{}.Find(expect.Id)
	assertError(t, err)
	assertEqualStruct(t, expect, u)
}

func TestFindBy(t *testing.T) {
	expect := &User{Name: "test"}
	expect.Save()
	defer User{}.DeleteAll()

	u, err := User{}.FindBy("name", "test")
	assertError(t, err)
	assertEqualStruct(t, expect, u)
}

func TestFirst(t *testing.T) {
	for _, name := range []string{"test1", "test2"} {
		u := &User{Name: name}
		u.Save()
	}
	defer User{}.DeleteAll()

	expect, _ := User{}.Where("name", "test1").QueryRow()

	u, err := User{}.First()
	assertError(t, err)
	assertEqualStruct(t, expect, u)
}

func TestLast(t *testing.T) {
	for _, name := range []string{"test1", "test2"} {
		u := &User{Name: name}
		u.Save()
	}
	defer User{}.DeleteAll()

	expect, _ := User{}.Where("name", "test2").QueryRow()

	u, err := User{}.Last()
	assertError(t, err)
	assertEqualStruct(t, expect, u)
}

func TestWhere(t *testing.T) {
	expect := &User{Name: "test"}
	expect.Save()
	defer User{}.DeleteAll()

	u, err := User{}.Where("name", "test").And("id", expect.Id).QueryRow()

	assertError(t, err)
	assertEqualStruct(t, expect, u)
}

func TestOrder(t *testing.T) {
	expects := []string{"test1", "test2"}
	for _, name := range expects {
		u := &User{Name: name}
		u.Save()
	}
	defer User{}.DeleteAll()

	users, err := User{}.Order("name", "ASC").Query()

	assertError(t, err)
	for i, u := range users {
		if u.Name != expects[i] {
			t.Errorf("column value should be %v, but %v", expects[i], u.Name)
		}
	}
}

func TestLimitAndOffset(t *testing.T) {
	for _, name := range []string{"test1", "test2", "test3"} {
		u := &User{Name: name}
		u.Save()
	}
	defer User{}.DeleteAll()

	users, err := User{}.Limit(2).Offset(1).Order("name", "ASC").Query()

	assertError(t, err)
	expects := []string{"test2", "test3"}
	for i, u := range users {
		if u.Name != expects[i] {
			t.Errorf("column value should be %v, but %v", expects[i], u.Name)
		}
	}
}

func TestGroupByAndHaving(t *testing.T) {
	for _, name := range []string{"testA", "testB", "testB"} {
		u := &User{Name: name}
		u.Save()
	}
	defer User{}.DeleteAll()

	users, err := User{}.Group("name").Having("count(name)", 2).Query()

	assertError(t, err)
	expects := []string{"testB"}
	for i, u := range users {
		if u.Name != expects[i] {
			t.Errorf("column value should be %v, but %v", expects[i], u.Name)
		}
	}
}

func TestExplain(t *testing.T) {
	err := User{}.Where("name", "test").Explain()
	assertError(t, err)
}

func TestIsValid(t *testing.T) {
	p := &Post{Name: "abc"}
	_, errs := p.IsValid()

	if len(errs.Messages["name"]) != 1 {
		t.Errorf("errors count should be 1, but %d", len(errs.Messages["name"]))
	}
}

func TestBuild(t *testing.T) {
	defer User{}.DeleteAll()
	u, _ := User{}.Create(UserParams{Name: "TestCreate"})

	p := u.BuildPost(PostParams{Name: "name"})

	if p.UserId != u.Id {
		t.Errorf("column value should be %v, but %v", u.Id, p.UserId)
	}
}

func TestCreate(t *testing.T) {
	u, errs := User{}.Create(UserParams{
		Name: "TestCreate",
	})
	defer User{}.DeleteAll()

	assertError(t, errs)

	expect, _ := User{}.FindBy("name", "TestCreate")
	assertEqualStruct(t, expect, u)
}

func TestIsNewRecordAndIsPresistent(t *testing.T) {
	defer User{}.DeleteAll()

	u := &User{Name: "test"}
	if !u.IsNewRecord() {
		t.Errorf("struct is new record, but isn't new record.")
	}

	u.Save()
	if !u.IsPersistent() {
		t.Errorf("struct is persistent, but isn't persistent.")
	}
}

func TestSaveWithInvalidData(t *testing.T) {
	defer Post{}.DeleteAll()

	// OnCreate
	p := &Post{Name: "invalid"}
	_, errs := p.Save()

	if len(errs.Messages["name"]) != 1 {
		t.Errorf("errors count should be 1, but %d", len(errs.Messages["name"]))
	}

	p.Name = "name"
	_, errs = p.Save()
	assertError(t, errs)

	// OnUpdate
	p.Name = "invalid2"
	_, errs = p.Save()

	if len(errs.Messages["name"]) != 1 {
		t.Errorf("errors count should be 1, but %d", len(errs.Messages["name"]))
	}

}

func TestSave(t *testing.T) {
	defer User{}.DeleteAll()

	u := &User{Name: "test"}

	_, errs := u.Save()
	assertError(t, errs)

	if u.Id == 0 {
		t.Errorf("Id should be setted after save, but isn't setted")
	}

	expect, _ := User{}.FindBy("name", "test")
	assertEqualStruct(t, expect, u)

	u.Name = "test2"
	_, errs = u.Save()
	assertError(t, errs)

	expect, _ = User{}.Find(u.Id)
	assertEqualStruct(t, expect, u)
}

func TestUpdate(t *testing.T) {
	defer User{}.DeleteAll()

	u := &User{Name: "test"}
	_, errs := u.Save()

	expect := UserParams{Name: "test2"}
	_, errs = u.Update(expect)
	assertError(t, errs)

	actual, _ := User{}.Find(u.Id)
	if expect.Name != actual.Name {
		t.Errorf("column value should be equal to %v, but %v", expect.Name, actual.Name)
	}
}

func TestUpdateColumns(t *testing.T) {
	defer User{}.DeleteAll()

	u := &User{Name: "test"}
	_, errs := u.Save()

	expect := UserParams{Name: "test2"}
	_, errs = u.UpdateColumns(expect)
	assertError(t, errs)

	actual, _ := User{}.Find(u.Id)
	if expect.Name != actual.Name {
		t.Errorf("column value should be equal to %v, but %v", expect.Name, actual.Name)
	}
}

func TestDelete(t *testing.T) {
	defer User{}.DeleteAll()

	u, _ := User{}.Create(UserParams{Name: "test1"})

	_, errs := u.Delete()
	assertError(t, errs)

	actual, _ := User{}.Find(u.Id)
	if actual != nil {
		t.Errorf("record should be deleted, but isn't deleted %v", actual)
	}
}

func TestScope(t *testing.T) {
	defer User{}.DeleteAll()

	User{}.Create(UserParams{Name: "test1", Age: 20})
	expect, _ := User{}.Create(UserParams{Name: "test2", Age: 21})

	users, _ := User{}.OlderThan(20).Query()

	if len(users) != 1 {
		t.Errorf("record count should be 1, but %v", len(users))
	}
	assertEqualStruct(t, users[0], expect)
}

func TestHasMany(t *testing.T) {
	defer func() {
		User{}.DeleteAll()
		Post{}.DeleteAll()
	}()

	u, _ := User{}.Create(UserParams{Name: "test1"})
	expect, _ := Post{}.Create(PostParams{UserId: u.Id, Name: "name"})

	user, err := User{}.Find(u.Id)
	posts, err := user.Posts()
	assertError(t, err)
	if len(posts) != 1 {
		t.Errorf("record count should be 1, but %v", len(posts))
	}
	assertEqualStruct(t, posts[0], expect)
}

func TestBelongsTo(t *testing.T) {
	defer func() {
		User{}.DeleteAll()
		Post{}.DeleteAll()
	}()

	expect, _ := User{}.Create(UserParams{Name: "test1"})
	p, _ := Post{}.Create(PostParams{UserId: expect.Id, Name: "name"})

	user, err := p.User()
	assertError(t, err)
	assertEqualStruct(t, user, expect)
}

func TestJoins(t *testing.T) {
	defer func() {
		User{}.DeleteAll()
		Post{}.DeleteAll()
	}()

	// User joins posts
	u, _ := User{}.Create(UserParams{Name: "test1"})
	p1, _ := Post{}.Create(PostParams{UserId: u.Id, Name: "name"})
	p2 := Post{UserId: u.Id, Name: "invalid"}
	p2.Save(false)

	users, err := User{}.JoinsPosts().Where("posts.name", "name").Query()
	assertError(t, err)
	if len(users) != 1 {
		t.Errorf("record count should be 1, but %v", len(users))
	}
	assertEqualStruct(t, users[0], u)

	// Posts joins user
	posts, err := Post{}.JoinsUser().Where("posts.name", "name").Query()
	assertError(t, err)

	if len(posts) != 1 {
		t.Errorf("record count should be 1, but %v", len(posts))
	}
	assertEqualStruct(t, posts[0], p1)
}

func TestExists(t *testing.T) {
	defer User{}.DeleteAll()
	exist := User{}.Exists()
	if exist {
		t.Errorf("record shouldn't exist, but exist")
	}

	User{}.Create(UserParams{Name: "test"})
	exist = User{}.Where("name", "test").Exists()
	if !exist {
		t.Errorf("record should exist, but dosen't exist")
	}
}

func TestCount(t *testing.T) {
	defer User{}.DeleteAll()
	count := User{}.Count()
	if count != 0 {
		t.Errorf("record count should be 0, but %v", count)
	}

	User{}.Create(UserParams{Name: "test"})
	count = User{}.Where("name", "test").Count("id")
	if count != 1 {
		t.Errorf("record count should be 1, but %v", count)
	}
}

func TestAll(t *testing.T) {
	defer User{}.DeleteAll()
	users, _ := User{}.All().Query()
	if len(users) != 0 {
		t.Errorf("record count should be 0, but %v", len(users))
	}

	User{}.Create(UserParams{Name: "test"})
	users, _ = User{}.Where("name", "test").All().Query()
	if len(users) != 1 {
		t.Errorf("record count should be 1, but %v", len(users))
	}
}

func assertEqualStruct(t *testing.T, expect, actual interface{}) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("struct should be equal to %v, but %v", expect, actual)
	}
}

func assertError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("error should be nil, but %v", err)
	}
}

func testDb() (*sql.DB, error) {
	switch os.Getenv("DB") {
	case "mysql":
		return sql.Open("mysql", "travis@/argen_test")
	case "sqlite3", "":
		return sql.Open("sqlite3", ":memory:")
	}
	return nil, nil
}

func testTables() []string {
	switch os.Getenv("DB") {
	case "mysql":
		return []string{
			"drop table if exists users;",
			"drop table if exists posts;",
			"create table users (id INTEGER PRIMARY KEY AUTO_INCREMENT, name text, age integer);",
			"create table posts (id INTEGER PRIMARY KEY AUTO_INCREMENT, user_id integer not null, name text);",
		}
	case "sqlite3", "":
		return []string{
			"create table users (id integer PRIMARY KEY AUTOINCREMENT, name text, age integer);",
			"create table posts (id integer PRIMARY KEY AUTOINCREMENT, user_id integer not null, name text);",
		}
	}
	return []string{}
}
