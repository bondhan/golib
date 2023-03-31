package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetTimeInLocation(t *testing.T) {
	assertTest := assert.New(t)

	utcTime := time.Date(2022, 11, 12, 00, 00, 00, 00, time.UTC)
	locationTest, _ := GetTimeInLocation(utcTime, "UTC")

	// test - no change
	assertTest.Equal(utcTime.Unix(), locationTest.Unix())
	assertTest.Equal(utcTime.Location(), time.UTC)

	// test - Jakarta
	locationTest, _ = GetTimeInLocation(utcTime, JakartaWibTz)
	assertTest.Equal(JakartaWibTz, locationTest.Location().String())

	// test - Jakarta
	locationTest, _ = GetTimeInLocation(utcTime, "Asia/Tokyo")
	assertTest.Equal("Asia/Tokyo", locationTest.Location().String())

}

func Test_GetTimeInJakartaLocation(t *testing.T) {
	assertTest := assert.New(t)

	utcTime := time.Date(2022, 11, 12, 00, 00, 00, 00, time.UTC)
	jakartaTime, _ := GetTimeInJakartaLocation(utcTime)

	assertTest.Equal("2022-11-12 07:00:00 +0700 WIB", jakartaTime.String())
}
func Test_DateStringToTime(t *testing.T) {
	assertTest := assert.New(t)
	actualTime, err := DateStringToTime("2022-03-05")
	require.NoError(t, err)
	assertTest.Equal("2022-03-05 00:00:00 +0000 UTC", actualTime.String())

	actualTime, err = DateStringToTime("2022-03-05T15:00:00")
	require.NoError(t, err)
	assertTest.Equal("2022-03-05 15:00:00 +0000 UTC", actualTime.String())

	actualTime, err = DateStringToTime("2022-03-05 15:00:00")
	require.NoError(t, err)
	assertTest.Equal("2022-03-05 15:00:00 +0000 UTC", actualTime.String())

	actualTime, err = DateStringToTime("2022-03-05T18:15:30+00:00")
	require.NoError(t, err)
	assertTest.Equal("2022-03-05 18:15:30 +0000 UTC", actualTime.UTC().String())

}

func Test_IsSameDate(t *testing.T) {
	var (
		deliveryStartAt int64
		deliveryEndAt   int64
		assertTest      = assert.New(t)
	)

	// test - same date
	deliveryStartAt = 1672275600
	deliveryEndAt = 1672286400
	res := IsSameDate(deliveryStartAt, deliveryEndAt)
	assertTest.True(res)

	// test - diffrent date
	deliveryStartAt = 1682275600
	deliveryEndAt = 1672286400
	res = IsSameDate(deliveryStartAt, deliveryEndAt)
	assertTest.False(res)
}
