package jsonwriter

import (
	"fmt"
	"io"
	"strconv"
	"time"
	"unicode/utf8"
	"encoding/json"
)

var (
	quote        = []byte(`"`)
	keyStart     = quote
	null         = []byte("null")
	_true        = []byte("true")
	_false       = []byte("false")
	comma        = []byte(",")
	keyEnd       = []byte(`":`)
	startObject  = []byte("{")
	endObject    = []byte("}")
	startArray   = []byte("[")
	endArray     = []byte("]")
	escapedQuote = []byte(`\"`)
	escapedSlash = []byte(`\\`)
	escapedBS    = []byte(`\b`)
	escapedFF    = []byte(`\f`)
	escapedNL    = []byte(`\n`)
	escapedLF    = []byte(`\r`)
	escapedTab   = []byte(`\t`)
)

type Writer struct {
	depth int
	first bool
	array bool
	W     io.Writer
}

// Creates a JsonWriter that writes to the provided io.Writer
func New(w io.Writer) *Writer {
	return &Writer{
		W:     w,
		first: true,
	}
}

// Starts the writing process by creating an object.
// Should only be called once
func (w *Writer) RootObject(f func()) {
	w.W.Write(startObject)
	f()
	w.W.Write(endObject)
}

// Starts the writing process by creating an array.
// Should only be called once
func (w *Writer) RootArray(f func()) {
	w.array = true
	w.W.Write(startArray)
	f()
	w.W.Write(endArray)
}

// Star an object with the given key
func (w *Writer) Object(key string, f func()) {
	w.Key(key)
	w.first = true
	w.W.Write(startObject)
	f()
	w.W.Write(endObject)
}

// Star an array with the given key
func (w *Writer) Array(key string, f func()) {
	w.Key(key)
	w.first, w.array = true, true
	w.W.Write(startArray)
	f()
	w.array = false
	w.W.Write(endArray)
}

// Star an object within an array (a keyless object)
func (w *Writer) ArrayObject(f func()) {
	w.first = true
	w.array = false
	w.W.Write(startObject)
	f()
	w.W.Write(endObject)
	w.array = true
}

// Writes a key. The key is placed within quotes and ends
// with a colon
func (w *Writer) Key(key string) {
	w.Separator()
	w.W.Write(keyStart)
	w.writeString(key)
	w.W.Write(keyEnd)
}

// value can be a string, byte, u?int(8|16|32|64)?, float(32|64)?,
// time.Time, bool, nil or encoding/json.Marshaller
func (w *Writer) Value(value interface{}) {
	if w.array {
		w.Separator()
	}

	if value == nil {
		w.W.Write(null)
		return
	}

	switch t := value.(type) {
	case bool:
		if t == true {
			w.W.Write(_true)
		} else {
			w.W.Write(_false)
		}
	case uint8:
		w.W.Write([]byte(strconv.FormatUint(uint64(t), 10)))
	case uint16:
		w.W.Write([]byte(strconv.FormatUint(uint64(t), 10)))
	case uint32:
		w.W.Write([]byte(strconv.FormatUint(uint64(t), 10)))
	case uint:
		w.W.Write([]byte(strconv.FormatUint(uint64(t), 10)))
	case uint64:
		w.W.Write([]byte(strconv.FormatUint(t, 10)))
	case int8:
		w.W.Write([]byte(strconv.FormatInt(int64(t), 10)))
	case int16:
		w.W.Write([]byte(strconv.FormatInt(int64(t), 10)))
	case int32:
		w.W.Write([]byte(strconv.FormatInt(int64(t), 10)))
	case int:
		w.W.Write([]byte(strconv.FormatInt(int64(t), 10)))
	case int64:
		w.W.Write([]byte(strconv.FormatInt(t, 10)))
	case float32:
		w.W.Write([]byte(strconv.FormatFloat(float64(t), 'g', -1, 32)))
	case float64:
		w.W.Write([]byte(strconv.FormatFloat(t, 'g', -1, 64)))
	case json.Marshaler:
		b, _ := t.MarshalJSON()
		w.W.Write(b)
	case string:
		w.W.Write(quote)
		w.writeString(t)
		w.W.Write(quote)
	case time.Time:
		w.W.Write(quote)
		w.writeString(t.Format(time.RFC3339Nano))
		w.W.Write(quote)
	default:
		panic(fmt.Sprintf("unsuported valued type %v", value))
	}
}

// writes a key: value
// This is the same as calling WriteKey(key) followe by WriteValue(value)
func (w *Writer) KeyValue(key string, value interface{}) {
	w.Key(key)
	w.Value(value)
}

func (w *Writer) Separator() {
	if w.first == false {
		w.W.Write(comma)
	} else {
		w.first = false
	}
}

func (w *Writer) writeString(s string) {
	start, end := 0, 0
	var special []byte
L:
	for i, r := range s {
		switch r {
		case '"':
			special = escapedQuote
		case '\\':
			special = escapedSlash
		case '\b':
			special = escapedBS
		case '\f':
			special = escapedFF
		case '\n':
			special = escapedNL
		case '\r':
			special = escapedLF
		case '\t':
			special = escapedTab
		default:
			end += utf8.RuneLen(r)
			continue L
		}

		if end > start {
			w.W.Write([]byte(s[start:end]))
		}
		w.W.Write(special)
		start = i + 1
		end = start
	}
	if end > start {
		w.W.Write([]byte(s[start:end]))
	}
}
