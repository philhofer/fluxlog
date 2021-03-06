// Flux/msg implements a stripped-down
// version of the MessagePack serialization protocol
// (http://msgpack.org/). Flux/msg does not prefix
// encodings with map or array headers; rather, it
// assumes that the sender and receiver share
// a "Schema" that defines the ordering, types,
// and names of each value in the message.
// Consequently, the messages themselves only
// contain a list of type identifiers and raw
// values, rather than an array or map. Write
// and Read methods still use the standard
// MessagePack single-byte type prefixes.
// Consider the following message:
//  {"compact":true, "schema":0}
// Ordinarily, messagepack would encode that as:
//  \0x82 \0xA7 compact \0xC3 \0xA6 schema \0x00
//  [2-element map][7-byte string]["compact"][true][6-byte string]["schema"][0 (int)]
// whereas Flux/msg encodes as:
//  \0xC3 \0x00
//  [true][0 (int)]
// Provided that both the sender and receiver know that this message
// type is always a boolean followed by an integer, and that the
// boolean is named "compact" and the integer is named "schema",
// the message above serves as a complete message. In that sense,
// we can consider these messages to be "strongly-typed," as they
// have a fixed ordering of non-optional named fields, and
// both senders and receivers must know the type specification
// beforehand. (To facilitate more runtime flexibility, Schema
// types know how to serialize and de-serialize themselves!)
// For small messages, Flux/msg is MUCH smaller than even
// standard messagepack (2 bytes vs 18 bytes in our example),
// and it encodes and decodes quickly because we
// avoid runtime type reflection.
package msg

// PackExt represents a MessagePack extension, and has msg.Type = msg.Ext.
// A messagepack extension is simply a tuple of an 8-bit type identifier with arbitary binary data.
type PackExt struct {
	// Type is an 8-bit signed integer. The MessagePack standard dictates that 0 through 127
	// are permitted, while negative values are reserved for future use.
	EType int8
	// Data is the data stored in the extension.
	Data []byte
}

/* Write takes an object and writes it to a Writer

Supported type-Type tuples are:

 - float64 - msg.Float
 - bool - msg.Bool
 - int64 - msg.Int
 - uint64 - msg.Uint
 - string - msg.String
 - []byte - msg.Bin
 - *msg.PackExt - msg.Ext (must be non-nil, otherwise panic)

Each type will be compacted on writing if it
does not require all of its bits to represent itself.
Write returns ErrTypeNotSupported if a bad type is given.
Write returns ErrIncorrectType if the type given does not match the interface{} type.
Alternatively, you can use one of the WriteXxxx() methods provided. */
func WriteInterface(w Writer, v interface{}, t Type) error {
	switch t {
	case String:
		s, ok := v.(string)
		if !ok {
			return ErrIncorrectType
		}
		writeString(w, s)
		return nil
	case Int:
		i, ok := v.(int64)
		if !ok {
			return ErrIncorrectType
		}
		writeInt(w, i)
		return nil
	case Uint:
		u, ok := v.(uint64)
		if !ok {
			return ErrIncorrectType
		}
		writeUint(w, u)
		return nil
	case Float:
		f, ok := v.(float64)
		if !ok {
			return ErrIncorrectType
		}
		writeFloat(w, f)
		return nil
	case Bool:
		t, ok := v.(bool)
		if !ok {
			return ErrIncorrectType
		}
		writeBool(w, t)
		return nil
	case Bin:
		b, ok := v.([]byte)
		if !ok {
			return ErrIncorrectType
		}
		writeBin(w, b)
		return nil
	case Ext:
		ext, ok := v.(*PackExt)
		if !ok {
			return ErrIncorrectType
		}
		writeExt(w, ext.EType, ext.Data)
		return nil
	default:
		return ErrTypeNotSupported
	}
}

//WriteFloat writes a float to a msg.Writer
func WriteFloat(w Writer, f float64) { writeFloat(w, f) }

//WriteFloat64 writes a float, but guarantees that the encoded value will be a full 64 bits
func WriteFloat64(w Writer, f float64) { writeFloat64(w, f) }

//WriteFloat32 writes a float, but guarantees that the encoded value will be 32 bits
func WriteFloat32(w Writer, f float32) { writeFloat32(w, f) }

//WriteBool writes a bool to a msg.Writer
func WriteBool(w Writer, b bool) { writeBool(w, b) }

//WriteInt writes an int to a msg.Writer
func WriteInt(w Writer, i int64) { writeInt(w, i) }

//WriteUint writes a uint to a msg.Writer
func WriteUint(w Writer, u uint64) { writeUint(w, u) }

//WriteString writes a string to a msg.Writer
func WriteString(w Writer, s string) { writeString(w, s) }

//WriteBin writes an arbitrary binary to a msg.Writer
func WriteBin(w Writer, b []byte) { writeBin(w, b) }

//WriteExt writes a messagepack 'extension' (tuple of type, data) to a msg.Writer
func WriteExt(w Writer, etype int8, data []byte) { writeExt(w, etype, data) }

// ReadXxxx() methods try to read values
// from a msg.Reader into a value.
// If the reader reads a leading tag that does not
// translate to the ReadXxxx() method called, it
// unreads the leading byte so that another
// ReadXxxx() method can be attempted and returns ErrBadTag.
//
// ReadFloat tries to read into a float64.
func ReadFloat(r Reader) (f float64, err error) {
	f, err = readFloat(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
	}
	return
}

// ReadFloat32 reads a float32
func ReadFloat32(r Reader) (f float32, err error) {
	f, err = readFloat32(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
	}
	return
}

// ReadFloat64 reads a float64
func ReadFloat64(r Reader) (f float64, err error) {
	f, err = readFloat64(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
	}
	return
}

// ReadFloatBytes returns a float64 from 'p' along with
// the number of bytes read, or an error.
func ReadFloatBytes(p []byte) (f float64, n int, err error) {
	return readFloatBytes(p)
}

// ReadFloat64Bytes attempts to read a float64 out of 'p'
func ReadFloat64Bytes(p []byte) (f float64, n int, err error) { return readFloat64Bytes(p) }

// ReadFloat32Bytes attempts to read a float32 out of 'p'
func ReadFloat32Bytes(p []byte) (f float32, n int, err error) { return readFloat32Bytes(p) }

// ReadInt tries to read into an int64
func ReadInt(r Reader) (i int64, err error) {
	i, err = readInt(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
	}
	return
}

// ReadIntBytes reads an int64 from 'p'
func ReadIntBytes(p []byte) (i int64, n int, err error) { return readIntBytes(p) }

// ReadUint tries to read into a uint64.
func ReadUint(r Reader) (u uint64, err error) {
	u, err = readUint(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
	}
	return
}

func ReadUintBytes(p []byte) (u uint64, n int, err error) { return readUintBytes(p) }

// ReadString tries to read into a string.
func ReadString(r Reader) (s string, err error) {
	s, err = readString(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
	}
	return
}

// ReadStringZeroCopy reads a string from 'p' into s.
// Note that the string returned violates immutability, as changes
// to 'p' will change 's'. (The underlying pointer-to-byte-array in
// 's' points to 'p'.) Use with extreme care!
func ReadStringZeroCopy(p []byte) (s string, n int, err error) { return readStringZeroCopy(p) }

// ReadBool tries to read into a bool.
func ReadBool(r Reader) (b bool, err error) {
	b, err = readBool(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
	}
	return
}

// ReadBoolBytes reads a bool from 'p', along with the number
// of bytes read, or an error.
func ReadBoolBytes(p []byte) (b bool, n int, err error) { return readBoolBytes(p) }

// ReadBin tries to read into a byte slice.
// The slice 'b' is used for buffering in order to avoid allocations,
// but it can safely be nil. Usually 'b' should be a slice
// of an array on the stack. ReadBin may use the entire underlying capacity of 'b',
// and 'b' and 'dat' may share memory. (In other words, data will be read into 'b'
// if it is large enough, in which case 'dat' will be a sub-slice of 'b'. If cap(dat)=cap(b),
// then they point to the same underlying array.)
func ReadBin(r Reader, b []byte) (dat []byte, err error) {
	dat, err = readBin(r, b)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
	}
	return
}

// ReadBinZeroCopy returns the slice of 'p' that corresponds to
// binary data, along with the number of bytes read, or an error.
func ReadBinZeroCopy(p []byte) (dat []byte, n int, err error) { return readBinZeroCopy(p) }

// ReadExt tries to read into an PackExt.
// The slice 'b' is used in order to avoid allocations,
// but it can safely be nil. In many cases, 'b' should be
// a slice on the stack that is re-used. 'b' and p.Data may share
// memory. ReadExt may use the entire underlying capacity of 'b'. See ReadBin.
func ReadExt(r Reader, b []byte) (p *PackExt, err error) {
	dat, etype, err := readExt(r, b)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
		return nil, err
	}
	p = &PackExt{EType: etype, Data: dat}
	return p, nil
}

// ReadExtZeroCopy returns the extension type and binary array
// starting at p[0:], along with the number of bytes read, or an error.
// Note that 'dat' is a slice of 'p' - changes to 'p' will be reflected in 'dat' and vice-versa.
func ReadExtZeroCopy(p []byte) (dat []byte, etype int8, n int, err error) { return readExtZeroCopy(p) }

// ReadInterface returns an interface{} containing the leading object in the reader,
// along with its msg.Type.
//
// Provided no error is returned, the following type assertions on the interface{} should be legal:
//  - msg.Int -> int64
//  - msg.Uint -> uint64
//  - msg.Bool -> bool
//  - msg.Ext -> *msg.PackExt
//  - msg.Bin -> []byte
//  - msg.String -> string
//  - msg.Float -> float64
func ReadInterface(r Reader) (v interface{}, t Type, err error) {
	var c byte

	c, err = r.ReadByte()
	if err != nil {
		return
	}

	//fixed encoding cases (fixint, nfixint, fixstr)
	switch {
	//fixint
	case (c & 0x80) == 0:
		t = Int
		v = int64(int8(c & 0x7f))
		return

	//negative fixint
	case (c & 0xe0) == 0xe0:
		t = Int
		v = int64(int8(c))
		return

	//fixstr
	case c&0xe0 == 0xa0:
		t = String
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v, err = readString(r)
		return
	}

	//non-fix cases
	switch c {
	case mfalse:
		t = Bool
		v = false
		return
	case mtrue:
		t = Bool
		v = true
		return
	case mint8, mint16, mint32, mint64:
		t = Int
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v, err = readInt(r)
		return
	case muint8, muint16, muint32, muint64:
		t = Uint
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v, err = readUint(r)
		return
	case mfloat32, mfloat64:
		t = Float
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v, err = readFloat(r)
		return
	case mbin8, mbin16, mbin32:
		t = Bin
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v, err = readBin(r, nil)
		return
	case mfixext1, mfixext2, mfixext4, mfixext8, mfixext16, mext8, mext16, mext32:
		t = Ext
		err = r.UnreadByte()
		if err != nil {
			return
		}
		var etype int8
		var dat []byte
		dat, etype, err = readExt(r, nil)
		if err != nil {
			return
		}
		v = &PackExt{EType: etype, Data: dat}
		return
	case mstr8, mstr16, mstr32:
		t = String
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v, err = readString(r)
		return
	default:
		err = ErrTypeNotSupported
		return
	}
}
