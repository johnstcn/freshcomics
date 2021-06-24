package ipinfo

import "net"

// DummyIPInfoer implements IPInfoer but actually does nothing
type DummyLookuper struct{}

func (l *DummyLookuper) Lookup(addr net.IP, result interface{}) error {
	return nil
}
