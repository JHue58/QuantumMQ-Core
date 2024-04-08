package core

import "strings"

// UniqueID 唯一身份ID，用于检索对应的MsgQueue
type UniqueID string

func (u UniqueID) String() string {
	return string(u)
}

func (u UniqueID) Split() (UniqueID, UniqueID) {
	sp := strings.Split(u.String(), ":")
	return UniqueID(sp[0]), UniqueID(sp[1])
}

func (u UniqueID) Join(id UniqueID) UniqueID {
	return UniqueID(u.String() + ":" + id.String())
}

var uniqueIDSet = generateUniqueIDSet()

func generateUniqueIDSet() map[string]UniqueID {
	// TODO 持久化
	return make(map[string]UniqueID)
}
