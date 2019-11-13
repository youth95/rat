package rat

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventEmitter_AddListener(t *testing.T) {
	a := assert.New(t)
	var emitter EventEmitter
	el := emitter.AddListener("f0", func(msg interface{}) error {
		return nil
	})
	a.NotNil(el)
}

func TestEventEmitter_Emit(t *testing.T) {
	a := assert.New(t)
	var emitter EventEmitter
	v := ""
	emitter.AddListener("f0", func(msg interface{}) error {
		v = msg.(string)
		return nil
	})
	has, err := emitter.Emit("f0", "a")
	a.Nil(err)
	a.True(has)
	a.Equal(v, "a")

	has, err = emitter.Emit("f1", "a")
	a.Nil(err)
	a.False(has)

	emitter.AddListener("f2", func(msg interface{}) error {
		v = msg.(string)
		return errors.New("error")
	})

	has, err = emitter.Emit("f2", "a")

	a.NotNil(err)
	a.False(has)
}

func TestEventEmitter_EventNames(t *testing.T) {
	a := assert.New(t)
	var emitter EventEmitter
	f := func(msg interface{}) error { return nil }

	emitter.AddListener("a", f)
	emitter.AddListener("a", f)
	emitter.AddListener("b", f)
	emitter.AddListener("c", f)

	a.NotEmpty(emitter.EventNames())
	a.Equal(len(emitter.EventNames()), 3)

}

func TestEventEmitter_ListenerCount(t *testing.T) {
	a := assert.New(t)
	var emitter EventEmitter
	f := func(msg interface{}) error { return nil }

	emitter.AddListener("a", f)
	emitter.AddListener("a", f)
	emitter.AddListener("b", f)
	emitter.AddListener("c", f)

	a.Equal(emitter.ListenerCount("a"), 2)
	a.Equal(emitter.ListenerCount("b"), 1)
	a.Equal(emitter.ListenerCount("c"), 1)
	a.Equal(emitter.ListenerCount("d"), 0)

}

func TestEventEmitter_Listeners(t *testing.T) {
	a := assert.New(t)

	var emitter EventEmitter
	f := func(msg interface{}) error { return nil }

	listeners := emitter.Listeners("a")

	a.Equal(len(listeners), 0)

	emitter.AddListener("a", f)

	listeners = emitter.Listeners("a")

	a.Equal(len(listeners), 1)
	a.Equal(listeners[0](nil), f(nil))

}

func TestEventEmitter_RemoveListener(t *testing.T) {
	a := assert.New(t)
	var emitter EventEmitter
	f := func(msg interface{}) error { return nil }

	h := emitter.AddListener("a", f)
	a.NotEmpty(emitter.ListenerCount("a"))
	emitter.RemoveListener("a", h)
	a.Empty(emitter.ListenerCount("a"))

	e := emitter.RemoveListener("a", h)
	a.NotNil(e)

	e = emitter.RemoveListener("b", h)
	a.NotNil(e)
}

func TestEventEmitter_Once(t *testing.T) {
	a := assert.New(t)
	var emitter EventEmitter
	count := 0
	f := func(msg interface{}) error {
		count += 1
		return nil
	}

	emitter.Once("a", f)

	ok, err := emitter.Emit("a", "")

	a.Nil(err)
	a.True(ok)
	a.Equal(count, 1)

	ok, err = emitter.Emit("a", "")

	a.Nil(err)
	a.False(ok)
	a.Equal(count, 1)

}

func TestEventEmitter_RemoveAllListener(t *testing.T) {
	a := assert.New(t)
	var emitter EventEmitter
	f := func(msg interface{}) error { return nil }

	emitter.AddListener("a", f)
	emitter.AddListener("a", f)
	emitter.AddListener("b", f)
	emitter.AddListener("c", f)

	emitter.RemoveAllListener("a")

	a.Equal(len(emitter.EventNames()), 2)

}
