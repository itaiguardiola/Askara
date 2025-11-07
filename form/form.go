package form

import "github.com/itaiguardiola/askara/errorlist"

type Form interface {
	Validate() errorlist.Errors
	String() string
}
