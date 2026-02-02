package providers

import "github.com/libdns/libdns"

type Adapter interface {
	libdns.RecordAppender
	libdns.RecordSetter
	libdns.RecordDeleter
}

type Provider interface {
	Name() string
	Type() string
	ZoneFilters() []string
	Adapter() Adapter
}
