package cron

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/robfig/cron"
	"github.com/stretchr/testify/assert"
)

func TestCron(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := cron.New()
		spec := "0 0 9-21 * * 1-5"
		assert.True(t, validateCronExpress(spec))
		parse, err := cron.Parse(spec)
		assert.Nil(t, err)
		t.Log(parse.Next(time.Now()))
		var i int
		assert.NoError(t, c.AddFunc(spec, func() {
			t.Log(parse.Next(time.Now()))
			i++
			if i == 10 {
				c.Stop()
			}
		}))
		c.Run()
		synctest.Wait()
	})
}
