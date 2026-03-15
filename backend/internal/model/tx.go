package model

// @Enum
type TxIsolationLevel uint8

const (
	TxIsolationLevelReadUncommitted TxIsolationLevel = iota
	TxIsolationLevelReadCommitted
	TxIsolationLevelRepeatableRead
	TxIsolationLevelSerializable
)

// @PublicValueInstance
type TxOptions struct {
	IsolationLevel TxIsolationLevel
	ReadOnly       bool
}
