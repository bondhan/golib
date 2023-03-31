package docstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bondhan/golib/cache"
	_ "github.com/bondhan/golib/cache/mem"
	"github.com/bondhan/golib/constant"
)

func TestDocstore(t *testing.T) {
	type User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Username  string    `json:"username"`
		Age       int       `json:"age"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	ms := NewMemoryStore("test", "id")
	cache, err := cache.New("mem://")
	require.Nil(t, err)
	conf := &Config{
		IDField:         defaultID,
		TimestampField:  defaultTimestamp,
		UpdateTimeField: "updated_at",
	}

	cs := NewDocstore(ms, cache, conf)
	assert.NotNil(t, cs)

	ctx := context.Background()

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

	assert.True(t, cache.Exist(ctx, usr.ID))

	doc.Age = 36
	require.Nil(t, cs.Update(ctx, doc))
	assert.False(t, cache.Exist(ctx, usr.ID))

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

	require.Nil(t, cs.Increment(ctx, nu.ID, "age", 2))
	require.Nil(t, cs.Get(ctx, nu.ID, &user))
	assert.Equal(t, 38, user.Age)

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
	count, err := cs.Count(ctx, q)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)

	q = &QueryOpt{
		Filter: []FilterOpt{
			{Field: "age", Ops: constant.GE, Value: 35},
		},
	}

	out = nil
	require.Nil(t, cs.Find(ctx, q, &out))
	assert.Equal(t, 5, len(out))
}

func TestBulkStore(t *testing.T) {
	type User struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Username  string    `json:"username"`
		Age       int       `json:"age"`
		CreatedAt time.Time `json:"created_at"`
	}

	ms := NewMemoryStore("test", "id")
	cache, err := cache.New("mem://")
	require.Nil(t, err)
	conf := &Config{
		IDField:        defaultID,
		TimestampField: defaultTimestamp,
	}

	cs := NewDocstore(ms, cache, conf)
	assert.NotNil(t, cs)

	ctx := context.Background()

	ins := make([]*User, 0)
	ids := make([]string, 0)
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

	require.Nil(t, cs.BulkCreate(ctx, ins, nil))

	var out []*User
	require.Nil(t, cs.BulkGet(ctx, ids, &out))

	for i := 0; i < 10; i++ {
		assert.Equal(t, ins[i].ID, out[i].ID)
		assert.Equal(t, ins[i].Name, out[i].Name)
		assert.Equal(t, ins[i].Age, out[i].Age)
		assert.Equal(t, ins[i].CreatedAt.Unix(), out[i].CreatedAt.Unix())
	}
}

func TestDocstoreTime(t *testing.T) {
	type User struct {
		ID        string `json:"id"`
		Name      string `json:"name,omitempty"`
		Username  string `json:"username,omitempty"`
		Age       int    `json:"age,omitempty"`
		CreatedAt int64  `json:"created_at,omitempty"`
		UpdatedAt int64  `json:"updated_at,omitempty"`
	}

	ms := NewMemoryStore("test", "id")
	cache, err := cache.New("mem://")
	require.Nil(t, err)
	conf := &Config{
		IDField:         defaultID,
		TimestampField:  defaultTimestamp,
		UpdateTimeField: "updated_at",
	}

	cs := NewDocstore(ms, cache, conf)
	assert.NotNil(t, cs)

	ctx := context.Background()

	usr := &User{
		Name:     "sahal",
		Username: "sahalzain",
		Age:      35,
	}

	require.Nil(t, cs.Create(ctx, usr))
	tn := time.Now()
	assert.NotEmpty(t, usr.ID)
	assert.Equal(t, tn.UnixMilli(), usr.CreatedAt)

	var doc User
	require.Nil(t, cs.Get(ctx, usr.ID, &doc))
	assert.Equal(t, usr.ID, doc.ID)
	assert.Equal(t, usr.Name, doc.Name)
	assert.Equal(t, usr.Username, doc.Username)
	assert.Equal(t, usr.Age, doc.Age)
	assert.Equal(t, tn.UnixMilli(), usr.CreatedAt)

	nusr := &User{
		ID:  usr.ID,
		Age: 38,
	}
	require.Nil(t, cs.Update(ctx, nusr))

	var ndoc User
	require.Nil(t, cs.Get(ctx, usr.ID, &ndoc))
	assert.Equal(t, usr.CreatedAt, ndoc.CreatedAt)
	assert.Equal(t, usr.CreatedAt, ndoc.CreatedAt)
	assert.Equal(t, usr.Name, ndoc.Name)
	assert.Equal(t, usr.Username, ndoc.Username)
	assert.Equal(t, 38, ndoc.Age)
}

func TestFindOne(t *testing.T) {
	type User struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Username  string `json:"username"`
		Age       int    `json:"age"`
		CreatedAt int64  `json:"created_at"`
	}

	ms := NewMemoryStore("test", "id")
	cache, err := cache.New("mem://")
	require.Nil(t, err)
	conf := &Config{
		IDField:        defaultID,
		TimestampField: defaultTimestamp,
	}

	cs := NewDocstore(ms, cache, conf)
	assert.NotNil(t, cs)

	ctx := context.Background()

	now := time.Now().UnixMilli()
	usr := &User{
		Name:      "sahal",
		Username:  "sahalzain",
		Age:       35,
		CreatedAt: now,
	}

	require.Nil(t, cs.Create(ctx, usr))
	assert.NotEmpty(t, usr.ID)
	assert.Equal(t, now, usr.CreatedAt)

	var doc User
	q := &QueryOpt{
		Filter: []FilterOpt{
			{Field: "username", Ops: constant.EQ, Value: "sahalzain"},
		},
	}
	require.Nil(t, cs.FindOne(ctx, q, &doc))

}
