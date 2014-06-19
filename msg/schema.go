package msg

import (
	"errors"
)

var (
	ErrTypeNotSupported = errors.New("Type not supported as Schema type") // ErrTypeNotSupported returns when creating a schema with an interface{} of unsupported type
	ErrIncorrectType    = errors.New("Incorrect mapping of Type to type") // ErrIncorrectType is returned when value.(type) doesn't match msg.Type
	ErrBadArgs          = errors.New("Bad arguments.")                    // ErrBadArgs is returned when arguments are malformed.
)

//Schema represents an ordering of named objects
type Schema []Object

//Object represents a named object of known type
type Object struct {
	T    Type
	Name string
}

/* MakeSchema makes a Schema out of a []string and []interface{}.
The 'names' and 'types' slices *must* be the same length.
Supported interface{} values are:

 float64, float32
 uint8, uint16, uint32, uint64
 int8, int16, int32, int64
 bool
 string
 []byte (binary)

Note that even though MakeSchema accepts non-64-bit types, the types used in
Encode() *must* be 64-bit (float64, int64, uint64), because the interface{} is type-asserted
to those types internally. */
func MakeSchema(names []string, types []interface{}) (s *Schema, err error) {
	if len(names) != len(types) {
		err = ErrBadArgs
		return
	}
	o := make([]Object, len(names))

	for i, kind := range types {
		o[i].Name = names[i]
		switch kind.(type) {
		case float32, float64:
			o[i].T = Float
		case uint, uint8, uint16, uint32, uint64:
			o[i].T = Uint
		case int, int8, int16, int32, int64:
			o[i].T = Int
		case bool:
			o[i].T = Bool
		case string:
			o[i].T = String
		case []byte:
			o[i].T = Bin
		default:
			return nil, ErrTypeNotSupported
		}
	}
	s = (*Schema)(&o)
	return
}

// DecodeToMap uses a schema to decode a fluxmsg stream into a map[string]interface{}.
// The map keys are the Name fields of each msg.Object in the msg.Schema.
func (s *Schema) DecodeToMap(r Reader, m map[string]interface{}) error {
	var t Type
	var n string
	var ns interface{}
	var bs [16]byte
	var etype int8
	var dat []byte
	var err error
	for _, o := range *s {
		t = o.T
		n = o.Name
		switch t {

		case String:
			ns, err = readString(r)
			if err != nil {
				return err
			}
			m[n] = ns

		case Int:
			ns, err = readInt(r)
			if err != nil {
				return err
			}
			m[n] = ns

		case Uint:
			ns, err = readUint(r)
			if err != nil {
				return err
			}
			m[n] = ns

		case Float:
			ns, err = readFloat(r)
			if err != nil {
				return err
			}
			m[n] = ns

		case Bin:
			dat, err = readBin(r, bs[:16])
			if err != nil {
				return err
			}
			m[n] = dat

		case Ext:
			dat, etype, err = readExt(r, bs[:16])
			if err != nil {
				return err
			}
			m[n] = &PackExt{Type: etype, Data: dat}

		default:
			err = ErrIncorrectType
			return err

		}
	}
	return nil
}

//Encode uses a schema to encode a slice-of-interface to a writer.
func (s *Schema) Encode(a []interface{}, w Writer) (err error) {
	for i, v := range a {
		err = encode(v, (*s)[i], w)
		if err != nil {
			return
		}
	}
	return
}

func encode(v interface{}, o Object, w Writer) error {
	switch o.T {
	case Float:
		f, ok := v.(float64)
		if !ok {
			return ErrIncorrectType
		}
		writeFloat(w, f)
		return nil
	case Uint:
		i, ok := v.(uint64)
		if !ok {
			return ErrIncorrectType
		}
		writeUint(w, i)
		return nil
	case Int:
		i, ok := v.(int64)
		if !ok {
			return ErrIncorrectType
		}
		writeInt(w, i)
		return nil
	case Bool:
		b, ok := v.(bool)
		if !ok {
			return ErrIncorrectType
		}
		writeBool(w, b)
		return nil
	case String:
		s, ok := v.(string)
		if !ok {
			return ErrIncorrectType
		}
		writeString(w, s)
		return nil
	case Bin:
		bs, ok := v.([]byte)
		if !ok {
			return ErrIncorrectType
		}
		writeBin(w, bs)
		return nil
	default:
		return ErrTypeNotSupported
	}
}
