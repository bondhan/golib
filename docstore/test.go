package docstore

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bondhan/golib/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func DriverCRUDTest(d Driver, t *testing.T) {
	type User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Username  string    `json:"username"`
		Age       int       `json:"age"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at`
	}

	ctx := context.Background()
	ts := time.Now()
	usr := &User{
		ID:        "1234",
		Name:      "sahal",
		Username:  "sahalzain",
		Age:       35,
		CreatedAt: ts,
	}

	require.Nil(t, d.Create(ctx, usr))

	var doc User
	require.Nil(t, d.Get(ctx, usr.ID, &doc))
	assert.Equal(t, usr.ID, doc.ID)
	assert.Equal(t, usr.Name, doc.Name)
	assert.Equal(t, usr.Username, doc.Username)
	assert.Equal(t, usr.Age, doc.Age)
	assert.Equal(t, usr.CreatedAt.Unix(), doc.CreatedAt.Unix())

	doc.Age = 36
	require.Nil(t, d.Update(ctx, doc.ID, doc, false))

	var user User
	require.Nil(t, d.Get(ctx, usr.ID, &user))
	assert.Equal(t, doc.Age, user.Age)
	assert.Equal(t, ts.Unix(), user.CreatedAt.Unix())

	nu := &User{
		ID:   user.ID,
		Name: "Sahal Zain",
	}
	require.Nil(t, d.Update(ctx, nu.ID, nu, true))
	require.Nil(t, d.Get(ctx, nu.ID, &user))
	assert.Equal(t, 0, user.Age)

	require.Nil(t, d.UpdateField(ctx, nu.ID, []Field{{Name: "age", Value: 36}}))
	require.Nil(t, d.Get(ctx, nu.ID, &user))
	assert.Equal(t, 36, user.Age)

	require.Nil(t, d.Increment(ctx, nu.ID, "age", 1))
	require.Nil(t, d.Get(ctx, nu.ID, &user))
	assert.Equal(t, 37, user.Age)

	require.Nil(t, d.GetIncrement(ctx, nu.ID, "age", 1, &user))
	assert.Equal(t, 38, user.Age)

	require.Nil(t, d.Delete(ctx, nu.ID))
	require.NotNil(t, d.Get(ctx, nu.ID, &user))

	for i := 0; i < 10; i++ {
		u := &User{
			ID:        fmt.Sprintf("%v", i),
			Name:      "name" + fmt.Sprintf("%v", i),
			Age:       30 + i,
			CreatedAt: time.Now(),
		}
		require.Nil(t, d.Create(ctx, u))
	}

	q := &QueryOpt{
		Filter: []FilterOpt{
			{Field: "name", Ops: constant.EQ, Value: "name1"},
		},
	}

	var out []User

	require.Nil(t, d.Find(ctx, q, &out))
	assert.Equal(t, 1, len(out))

	count, err := d.Count(ctx, q)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)

	q = &QueryOpt{
		Filter: []FilterOpt{
			{Field: "age", Ops: constant.GE, Value: 35},
		},
	}

	out = nil
	require.Nil(t, d.Find(ctx, q, &out))
	assert.Equal(t, 5, len(out))

	q = &QueryOpt{
		Filter: []FilterOpt{
			{Field: "", Ops: constant.OR, Value: []FilterOpt{
				{
					Field: "name",
					Value: "1234",
					Ops:   constant.EQ,
				},
				{
					Field: "age",
					Value: 35,
					Ops:   constant.GE,
				},
			}},
		},
	}
	require.Nil(t, d.Find(ctx, q, &out))
	assert.Equal(t, 5, len(out))

	out = nil

	q = &QueryOpt{
		Filter: []FilterOpt{
			{Field: "age", Ops: constant.GE, Value: 32},
		},
		Limit:    5,
		Skip:     5,
		OrderBy:  "age",
		IsAscend: true,
	}

	out = nil
	require.Nil(t, d.Find(ctx, q, &out))
	assert.Equal(t, 3, len(out))
	assert.Equal(t, 37, out[0].Age)

}

func DriverBulkTest(d Driver, t *testing.T) {
	type User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Username  string    `json:"username"`
		Age       int       `json:"age"`
		CreatedAt time.Time `json:"created_at"`
	}
	ctx := context.Background()

	ins := make([]interface{}, 0)
	ids := make([]interface{}, 0)
	for i := 0; i < 10; i++ {

		u := &User{
			ID:        fmt.Sprintf("BLK-%v", i),
			Name:      "name" + fmt.Sprintf("%v", i),
			Age:       30 + i,
			CreatedAt: time.Now(),
		}
		ins = append(ins, u)
		ids = append(ids, fmt.Sprintf("BLK-%v", i))
	}

	require.Nil(t, d.BulkCreate(ctx, ins, nil))

	var out []*User
	require.Nil(t, d.BulkGet(ctx, ids, &out))

	require.Equal(t, 10, len(out))

	for i := 0; i < 10; i++ {
		assert.Equal(t, ins[i].(*User).ID, out[i].ID)
		assert.Equal(t, ins[i].(*User).Name, out[i].Name)
		assert.Equal(t, ins[i].(*User).Age, out[i].Age)
		assert.Equal(t, ins[i].(*User).CreatedAt.Unix(), out[i].CreatedAt.Unix())
	}
}

func DocstoreTestCRUD(cs *CachedStore, t *testing.T) {
	ctx := context.Background()

	t.Run("CRUD Test", func(t *testing.T) {
		type User struct {
			ID        string    `json:"id"`
			Name      string    `json:"name"`
			Username  string    `json:"username"`
			Age       int       `json:"age"`
			CreatedAt time.Time `json:"created_at"`
		}

		usr := &User{
			Name:     "sahal",
			Username: "sahalzain",
			Age:      35,
		}

		require.Nil(t, cs.Create(ctx, usr))

		assert.NotEmpty(t, usr.ID)
		assert.Equal(t, time.Now().Unix(), usr.CreatedAt.Unix())

		var doc User
		require.Nil(t, cs.Get(ctx, usr.ID, &doc))
		assert.Equal(t, usr.ID, doc.ID)
		assert.Equal(t, usr.Name, doc.Name)
		assert.Equal(t, usr.Username, doc.Username)
		assert.Equal(t, usr.Age, doc.Age)
		assert.Equal(t, usr.CreatedAt.Unix(), doc.CreatedAt.Unix())

		assert.True(t, cs.cache.Exist(ctx, usr.ID))

		doc.Age = 36
		require.Nil(t, cs.Update(ctx, doc))
		assert.False(t, cs.cache.Exist(ctx, usr.ID))

		var user User
		require.Nil(t, cs.Get(ctx, usr.ID, &user))
		assert.Equal(t, doc.Age, user.Age)
		assert.Equal(t, doc.CreatedAt.Unix(), user.CreatedAt.Unix())

		nu := &User{
			ID:   user.ID,
			Name: "Sahal Zain",
		}
		require.Nil(t, cs.Replace(ctx, nu))
		require.Nil(t, cs.Get(ctx, nu.ID, &user))
		assert.Equal(t, 0, user.Age)

		require.Nil(t, cs.UpdateField(ctx, nu.ID, "age", 36))
		require.Nil(t, cs.Get(ctx, nu.ID, &user))
		assert.Equal(t, 36, user.Age)

		var uf []Field
		uf = append(uf, Field{
			Name:  "name",
			Value: "Sahal",
		})
		uf = append(uf, Field{
			Name:  "age",
			Value: 34,
		})

		require.Nil(t, cs.UpdateFields(ctx, nu.ID, uf))
		require.Nil(t, cs.Get(ctx, nu.ID, &user))
		/*
			for _, row := range uf {

				fmt.Println(row)
				v := reflect.ValueOf(row)
				typeOfS := v.Type()
				for i := 0; i < v.NumField(); i++ {
					assert.Equal(t, v.Field(i).Interface(), typeOfS.Field(i).Name)
				}
			}
		*/
		assert.Equal(t, "Sahal", user.Name)
		assert.Equal(t, 34, user.Age)

		require.Nil(t, cs.Increment(ctx, nu.ID, "age", 2))
		require.Nil(t, cs.Get(ctx, nu.ID, &user))
		assert.Equal(t, 36, user.Age)

		require.Nil(t, cs.Delete(ctx, nu.ID))
		require.NotNil(t, cs.Get(ctx, nu.ID, &user))

		for i := 0; i < 10; i++ {
			u := &User{
				ID:        fmt.Sprintf("%v", i),
				Name:      "name" + fmt.Sprintf("%v", i),
				Age:       30 + i,
				CreatedAt: time.Now(),
			}
			require.Nil(t, cs.Create(ctx, u))
		}

		q := &QueryOpt{
			Filter: []FilterOpt{
				{Field: "name", Ops: constant.EQ, Value: "name1"},
			},
		}

		var out []User

		require.Nil(t, cs.Find(ctx, q, &out))
		assert.Equal(t, 1, len(out))

		q = &QueryOpt{
			Filter: []FilterOpt{
				{Field: "age", Ops: constant.GE, Value: 35},
			},
		}

		out = nil
		require.Nil(t, cs.Find(ctx, q, &out))
		assert.Equal(t, 5, len(out))
	})

	t.Run("UpdateMany Test", func(t *testing.T) {
		type Nested struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		type Obj struct {
			ID        string    `json:"id"`
			Nested    *Nested   `json:"nested"`
			Name      string    `json:"name"`
			CreatedAt time.Time `json:"created_at"`
		}

		nestedObjs := []Obj{
			{
				Nested: &Nested{
					ID:   "bar",
					Name: "baz",
				},
				Name: "qux",
			},
			{
				Nested: &Nested{
					ID:   "quux",
					Name: "baz",
				},
				Name: "jazz",
			},
		}

		for _, obj := range nestedObjs {
			obj := obj
			assert.Nil(t, cs.Create(ctx, &obj))
		}

		var list []Obj
		updateManyTests := []struct {
			filter []FilterOpt
			fields map[string]interface{}
			length int
			query  *QueryOpt
		}{
			{
				filter: []FilterOpt{
					{Field: "nested.name", Value: "baz", Ops: "$eq"},
					{Field: "name", Value: "jazz", Ops: "$eq"},
				},
				fields: map[string]interface{}{
					"nested.id": "zingg",
				},
				length: 1,
				query: &QueryOpt{Filter: []FilterOpt{{
					Field: "nested.name", Value: "baz", Ops: constant.EQ},
					{Field: "name", Value: "jazz", Ops: constant.EQ},
				}},
			},
			{
				filter: []FilterOpt{{Field: "nested.name", Value: "baz", Ops: "$eq"}},
				fields: map[string]interface{}{
					"nested.id": "spaam",
					"name":      "schtik",
				},
				length: 2,
				query:  &QueryOpt{Filter: []FilterOpt{{Field: "name", Value: "schtik", Ops: constant.EQ}}},
			},
		}

		for _, tt := range updateManyTests {
			err := cs.UpdateMany(ctx, tt.filter, tt.fields)
			if err != nil && errors.Is(err, OperationNotSupported) {
				t.SkipNow()
			}
			assert.NoError(t, err)

			err = cs.Find(ctx, tt.query, &list)
			assert.NoError(t, err)
			assert.Equal(t, len(list), tt.length)

			for _, v := range list {
				if tt.length == 1 {
					assert.Equal(t, v.Nested.ID, tt.fields["nested.id"])
					continue
				}
				assert.Equal(t, v.Name, tt.fields["name"])
				assert.Equal(t, v.Nested.ID, tt.fields["nested.id"])
			}
			list = nil
		}
	})

}
