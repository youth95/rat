package rat

import (
	"container/list"
	"sync"
)

type EventEmitter struct {
	listenerMapper sync.Map // map[string]*List
}

type Listener func(msg interface{}) error

func (eventEmitter *EventEmitter) AddListener(eventName interface{}, listener Listener) *list.Element {
	v, _ := eventEmitter.listenerMapper.LoadOrStore(eventName, list.New())
	listeners := v.(*list.List)
	return listeners.PushBack(listener)
}

func (eventEmitter *EventEmitter) On(eventName interface{}, listener Listener) *list.Element {
	return eventEmitter.AddListener(eventName, listener)
}

func (eventEmitter *EventEmitter) Emit(eventName interface{}, msg interface{}) (bool, error) {
	v, _ := eventEmitter.listenerMapper.Load(eventName)
	if v == nil {
		return false, nil
	}
	listeners := v.(*list.List)
	if listeners.Len() == 0 {
		return false, nil
	}
	el := listeners.Front()
	for el != nil {
		err := el.Value.(Listener)(msg)
		if err != nil {
			return false, err
		}
		el = el.Next()
	}
	return true, nil
}

func (eventEmitter *EventEmitter) EventNames() []interface{} {
	var result []interface{}
	eventEmitter.listenerMapper.Range(func(key, value interface{}) bool {
		result = append(result, key.(string))
		return true
	})
	return result
}

func (eventEmitter *EventEmitter) ListenerCount(eventName interface{}) int {
	v, _ := eventEmitter.listenerMapper.Load(eventName)
	if v == nil {
		return 0
	}
	listeners := v.(*list.List)
	return listeners.Len()
}

func (eventEmitter *EventEmitter) Listeners(eventName interface{}) []Listener {
	var result []Listener
	v, _ := eventEmitter.listenerMapper.Load(eventName)
	if v == nil {
		return result
	}
	listeners := v.(*list.List)
	el := listeners.Front()
	for el != nil {
		result = append(result, el.Value.(Listener))
		el = el.Next()
	}
	return result
}

func (eventEmitter *EventEmitter) RemoveListener(eventName interface{}, el *list.Element) *EventEmitter {
	v, _ := eventEmitter.listenerMapper.Load(eventName)
	if v == nil {
		return eventEmitter
	}
	listeners := v.(*list.List)
	listeners.Remove(el)
	return eventEmitter
}

func (eventEmitter *EventEmitter) RemoveAllListener(eventName interface{}) *EventEmitter {
	eventEmitter.listenerMapper.Delete(eventName)
	return eventEmitter
}

func (eventEmitter *EventEmitter) Once(eventName interface{}, listener Listener) *list.Element {
	var el *list.Element
	el = eventEmitter.AddListener(eventName, func(msg interface{}) error {
		eventEmitter.RemoveListener(eventName, el)
		return listener(msg)
	})
	return el
}
