package config

import (
	"fmt"
	"github.com/teocci/go-rtsp-nvr/src/session"
	"reflect"
	"strconv"
	"strings"
)

const (
	formatTagValue = "%s=%s"
)

type Local struct {
	Server  session.Server
	Streams map[string]session.Stream
}

func (l *Local) TagValues() (names []string) {
	rt := reflect.TypeOf(*l)
	rv := reflect.ValueOf(l).Elem()

	names = make([]string, rt.NumField())
	for i := range names {

		tf := rt.Field(i)
		vf := rv.Field(i)
		k := tf.Tag.Get("json")
		val := reflect.ValueOf(vf.Interface())
		var v string
		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v = strconv.FormatInt(val.Int(), 10)
		case reflect.String:
			v = val.String()
		}

		names[i] = fmt.Sprintf(formatTagValue, k, v)
	}

	return
}

func (l *Local) TagNames() (names []string) {
	rt := reflect.TypeOf(*l)

	names = make([]string, rt.NumField())
	for i := range names {
		f := rt.Field(i)
		names[i] = f.Tag.Get("json")
	}

	return
}

func (l *Local) FieldNames() (names []string) {
	rt := reflect.TypeOf(*l)

	names = make([]string, rt.NumField())
	for i := range names {
		f := rt.Field(i)
		names[i] = f.Name
	}

	return
}

func (l *Local) StringConf() string {
	return strings.Join(l.TagValues(), " ")
}

func (l *Local) IsEmpty() bool {
	return &Local{} == l
}
