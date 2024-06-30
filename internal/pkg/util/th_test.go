package util_test

import (
	"go-todolist-grpc/internal/pkg/util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetServiceTimeLoc(t *testing.T) {
	loc := util.GetServiceTimeLoc()
	assert.Equal(t, "GMT+8", loc.String())
}

func TestGetFullDateStr(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		dateStr := util.GetFullDateStr(now)
		assert.NotEmpty(t, dateStr)
	})
}

func TestGetSimpleDateStr(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		dateStr := util.GetSimpleDateStr(now)
		assert.NotEmpty(t, dateStr)
	})
}

func TestGetDateStr(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		dateStr := util.GetDateStr(now)
		assert.NotEmpty(t, dateStr)
	})
}

func TestGetTimeValue(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		dateStr, timestamp := util.GetTimeValue(now)
		assert.NotEmpty(t, dateStr)
		assert.NotEqual(t, int64(0), timestamp)
	})
}

func TestGetTimeValuePtr(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		dateStrPtr, timestampPtr := util.GetTimeValuePtr(&now)
		assert.NotNil(t, dateStrPtr)
		assert.NotNil(t, timestampPtr)
	})

	t.Run("Failure_NilTime", func(t *testing.T) {
		dateStrPtr, timestampPtr := util.GetTimeValuePtr(nil)
		assert.Nil(t, dateStrPtr)
		assert.Nil(t, timestampPtr)
	})
}

func TestGetTimestampMS(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		timestamp := util.GetTimestampMS(now)
		assert.NotEqual(t, int64(0), timestamp)
	})
}

func TestCalculateEndTimeOfDay(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		endTime := util.CalculateEndTimeOfDay(now)
		assert.Equal(t, 23, endTime.Hour())
		assert.Equal(t, 59, endTime.Minute())
		assert.Equal(t, 59, endTime.Second())
	})
}

func TestMsTimestampStrToTime(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		timestamp := "1685106633000"
		result := util.MsTimestampStrToTime(timestamp)
		assert.NotNil(t, result)
	})

	t.Run("Failure_InvalidFormat", func(t *testing.T) {
		timestamp := "invalid"
		result := util.MsTimestampStrToTime(timestamp)
		assert.Nil(t, result)
	})
}
