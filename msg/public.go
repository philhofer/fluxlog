package msg

// PackExt represents a MessagePack extension
type PackExt struct {
	// Type is an 8-bit signed integer. The MessagePack standard dictates that 0 through 127
	// are permitted, while negative values are reserved for future use.
	Type int8
	// Data is the data stored in the extension.
	Data []byte
}

// Write takes an object and writes it to a Writer
// Supported type are:
// - float64 (msg.Float)
// - bool (msg.Bool)
// - int64 (msg.Int)
// - uint64 (msg.Uint)
// - string (msg.String)
// - []byte (msg.Bin)
// - msg.Ext (msg.Ext) - a messagepack extension type
// Each type will be compacted on writing if it
// does not require all of its bits to represent itself.
// Write returns ErrTypeNotSupported if a bad type is given.
// Write returns ErrIncorrectType if the type given does not match the interface{} type
func Write(w Writer, v interface{}, t Type) error {
	switch t {
	case String:
		s, ok := v.(string)
		if !ok {
			return ErrIncorrectType
		}
		writeString(w, s)
	case Int:
		i, ok := v.(int64)
		if !ok {
			return ErrIncorrectType
		}
		writeInt(w, i)
	case Uint:
		u, ok := v.(uint64)
		if !ok {
			return ErrIncorrectType
		}
		writeUint(w, u)
	case Float:
		f, ok := v.(float64)
		if !ok {
			return ErrIncorrectType
		}
		writeFloat(w, f)
	case Bool:
		t, ok := v.(bool)
		if !ok {
			return ErrIncorrectType
		}
		writeBool(w, t)
	case Bin:
		b, ok := v.([]byte)
		if !ok {
			return ErrIncorrectType
		}
		writeBin(w, b)
	case Ext:
		ext, ok := v.(PackExt)
		if !ok {
			return ErrIncorrectType
		}
		writeExt(w, ext.Type, ext.Data)
	default:
		return ErrTypeNotSupported
	}
	return ErrTypeNotSupported
}

// ReadXxxx() methods try to read values
// from a msg.Reader into a pointer-to-type.
// If the reader reads a leading tag that does not
// translate to the ReadXxxx() method called, it
// unreads the leading byte so that another
// ReadXxxx() method can be attempted.
//
//

// ReadFloat tries to read into a float64
func ReadFloat(r Reader, f *float64) error {
	g, err := readFloat(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
		return err
	}
	f = &g
	return nil
}

// ReadInt tries to read into an int64
func ReadInt(r Reader, i *int64) error {
	g, err := readInt(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
		return err
	}
	i = &g
	return nil
}

// ReadUint tries to read into a uint64
func ReadUint(r Reader, u *uint64) error {
	g, err := readUint(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
		return err
	}
	u = &g
	return nil
}

// ReadString tries to read into a string
func ReadString(r Reader, s *string) error {
	g, err := readString(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
		return err
	}
	s = &g
	return nil
}

// ReadBool tries to read into a bool
func ReadBool(r Reader, b *bool) error {
	g, err := readBool(r)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
		return err
	}
	b = &g
	return nil
}

// ReadBin tries to read into a byte slice
func ReadBin(r Reader, b []byte) error {
	err := readBin(r, b)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
		return err
	}
	return nil
}

// ReadExt tries to read into an PackExt
func ReadExt(r Reader, e *PackExt) error {
	etype, err := readExt(r, e.Data)
	if err != nil {
		if err == ErrBadTag {
			r.UnreadByte()
		}
		return err
	}
	e.Type = etype
	return nil
}

// ReadInterface returns an interface containing the leading object in the reader
// NOTE: Reading an interface value and type-switching on it eliminates
// most of the performance advantages of fluxmsg encoding, so only
// use it if you absolutely have to.
func ReadInterface(r Reader) (v interface{}, err error) {
	var c byte

	c, err = r.ReadByte()
	if err != nil {
		return
	}

	//fixed encoding cases (fixint, nfixint, fixstr)
	switch {
	//fixint
	case (c & 0x80) == 0:
		v = int64(int8(c & 0x7f))
		return

	//negative fixint
	case (c & 0xe0) == 0xe0:
		v = int64(int8(c))
		return

	//fixstr
	case c&0xe0 == 0xa0:
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
		v = false
		return
	case mtrue:
		v = true
		return
	case mint8, mint16, mint32, mint64:
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v, err = readInt(r)
		return
	case muint8, muint16, muint32, muint64:
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v, err = readUint(r)
		return
	case mfloat32, mfloat64:
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v, err = readFloat(r)
		return
	case mbin8, mbin16, mbin32:
		err = r.UnreadByte()
		if err != nil {
			return
		}
		v = make([]byte, 0, 32)
		err = readBin(r, v.([]byte))
		return
	case mfixext1, mfixext2, mfixext4, mfixext8, mfixext16, mext8, mext16, mext32:
		err = r.UnreadByte()
		if err != nil {
			return
		}
		var etype int8
		data := make([]byte, 0, 32)
		etype, err = readExt(r, data)
		v = &PackExt{Type: etype, Data: data}
		return
	case mstr8, mstr16, mstr32:
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