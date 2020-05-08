package filters

import (
	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
	"testing"
)

func TestNewHashFilter(t *testing.T) {
	filter, err := NewHashFilter(map[string]string{"none": "none", "md5": "false", "sha256": "false", "target": "foo"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Hash); ok {
		if e.useMd5 == true {
			t.Errorf("'md5' parameter ignored")
		}
		if e.useSha256 == true {
			t.Errorf("'sha256' parameter ignored")
		}
		if e.target != "foo" {
			t.Errorf("'target' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestHashDoFilterNoExtract(t *testing.T) {
	filter, err := NewHashFilter(map[string]string{"md5": "true", "sha1": "false"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Hash); ok {
		msg := "md5: 20a2fbdd6bd75fab08b9049da9c4ffe6 none 3cd8f3c8afa26f16fa4a5e3c8b56d703cf04135dafd7d8e2dd9f9433982abe46 | 072547a97a9efc2fad20b906d7e6697931c388a1"
		m := data.NewMessage(msg)
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == false {
			t.Errorf("it should return true")
		}
		if m.GetMessage() != msg {
			t.Errorf("the message has been altered by the filter")
		}

		msg = "aaaaaaa 072547a97a9efc2fad20b906d7e6697931c388a1"
		m = data.NewMessage(msg)
		b, err = e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return false")
		}
		if m.GetMessage() != msg {
			t.Errorf("the message has been altered by the filter")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestHashDoFilterExtractMd5(t *testing.T) {
	filter, err := NewHashFilter(map[string]string{
		"md5":     "true",
		"sha1":    "false",
		"sha256":  "false",
		"sha512":  "false",
		"extract": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	fb := NewFakeBus()

	filter.setBus(EventBus.Bus(fb))
	if e, ok := filter.(*Hash); ok {
		m := data.NewMessage("md5: 20a2fbdd6bd75fab08b9049da9c4ffe6 none 6b106759fa7fb86bdb60d0c5d77356adb5c365e2fdd781ab5bcc9645bb10e100 | 072547a97a9efc2fad20b906d7e6697931c388a1")
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return true")
		}

		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "20a2fbdd6bd75fab08b9049da9c4ffe6" {
			t.Errorf("hash has not been extracted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestHashDoFilterExtractSha1(t *testing.T) {
	filter, err := NewHashFilter(map[string]string{
		"md5":     "false",
		"sha1":    "true",
		"sha256":  "false",
		"sha512":  "false",
		"extract": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	fb := NewFakeBus()

	filter.setBus(EventBus.Bus(fb))
	if e, ok := filter.(*Hash); ok {
		m := data.NewMessage("md5: 20a2fbdd6bd75fab08b9049da9c4ffe6 none 6b106759fa7fb86bdb60d0c5d77356adb5c365e2fdd781ab5bcc9645bb10e100 | 072547a97a9efc2fad20b906d7e6697931c388a1")
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return true")
		}

		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "072547a97a9efc2fad20b906d7e6697931c388a1" {
			t.Errorf("hash has not been extracted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestHashDoFilterExtractSha256(t *testing.T) {
	filter, err := NewHashFilter(map[string]string{
		"md5":     "false",
		"sha1":    "false",
		"sha256":  "true",
		"sha512":  "false",
		"extract": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	fb := NewFakeBus()

	filter.setBus(EventBus.Bus(fb))
	if e, ok := filter.(*Hash); ok {
		m := data.NewMessage("md5: 20a2fbdd6bd75fab08b9049da9c4ffe6 none 6b106759fa7fb86bdb60d0c5d77356adb5c365e2fdd781ab5bcc9645bb10e100 | 072547a97a9efc2fad20b906d7e6697931c388a1")
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return true")
		}

		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "6b106759fa7fb86bdb60d0c5d77356adb5c365e2fdd781ab5bcc9645bb10e100" {
			t.Errorf("hash has not been extracted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestHashDoFilterExtractSha512(t *testing.T) {
	filter, err := NewHashFilter(map[string]string{
		"md5":     "false",
		"sha1":    "false",
		"sha256":  "false",
		"sha512":  "true",
		"extract": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	fb := NewFakeBus()

	filter.setBus(EventBus.Bus(fb))
	if e, ok := filter.(*Hash); ok {
		m := data.NewMessage("md5: 20a2fbdd6bd75fab08b9049da9c4ffe6 none 6b106759fa7fb86bdb60d0c5d77356adb5c365e2fdd781ab5bcc9645bb10e100 | 072547a97a9efc2fad20b906d7e6697931c388a1 2D58BA046462BE04415EE0D39243C6893C36F3AE438B5B20FA32E5596227EAD4C94599081ACD1C603AAD80088EB58B795683DDE7E14C579E4BDAE820E3628C33")
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return true")
		}

		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "2D58BA046462BE04415EE0D39243C6893C36F3AE438B5B20FA32E5596227EAD4C94599081ACD1C603AAD80088EB58B795683DDE7E14C579E4BDAE820E3628C33" {
			t.Errorf("hash has not been extracted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestHashDoFilterExtractAll(t *testing.T) {
	filter, err := NewHashFilter(map[string]string{
		"extract": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	fb := NewFakeBus()

	filter.setBus(EventBus.Bus(fb))
	if e, ok := filter.(*Hash); ok {
		m := data.NewMessage("md5: 20a2fbdd6bd75fab08b9049da9c4ffe6 none 6b106759fa7fb86bdb60d0c5d77356adb5c365e2fdd781ab5bcc9645bb10e100 | 072547a97a9efc2fad20b906d7e6697931c388a1 2D58BA046462BE04415EE0D39243C6893C36F3AE438B5B20FA32E5596227EAD4C94599081ACD1C603AAD80088EB58B795683DDE7E14C579E4BDAE820E3628C33")
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return true")
		}

		expected := []string{
			"20a2fbdd6bd75fab08b9049da9c4ffe6",
			"6b106759fa7fb86bdb60d0c5d77356adb5c365e2fdd781ab5bcc9645bb10e100",
			"072547a97a9efc2fad20b906d7e6697931c388a1",
			"2D58BA046462BE04415EE0D39243C6893C36F3AE438B5B20FA32E5596227EAD4C94599081ACD1C603AAD80088EB58B795683DDE7E14C579E4BDAE820E3628C33",
		}

		for i, v := range fb.Collected {
			if v.GetMessage() != expected[i] {
				t.Errorf("hash has not been extracted correctly: expected %s had %s", expected[i], v.GetMessage())
			}
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

type FakeBus struct {
	Collected []*data.Message
}

func NewFakeBus() *FakeBus {
	return &FakeBus{
		Collected: make([]*data.Message, 0),
	}
}

func (b *FakeBus) Reset() {
	b.Collected = make([]*data.Message, 0)
}

func (b *FakeBus) Publish(topic string, args ...interface{}) {
	for _, k := range args {
		if v, ok := k.(*data.Message); ok {
			b.Collected = append(b.Collected, v)
		}
	}
}

func (b *FakeBus) HasCallback(topic string) bool                                         { return true }
func (b *FakeBus) WaitAsync()                                                            {}
func (b *FakeBus) Subscribe(topic string, fn interface{}) error                          { return nil }
func (b *FakeBus) SubscribeAsync(topic string, fn interface{}, transactional bool) error { return nil }
func (b *FakeBus) SubscribeOnce(topic string, fn interface{}) error                      { return nil }
func (b *FakeBus) SubscribeOnceAsync(topic string, fn interface{}) error                 { return nil }
func (b *FakeBus) Unsubscribe(topic string, handler interface{}) error                   { return nil }
