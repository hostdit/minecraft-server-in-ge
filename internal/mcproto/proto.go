package mcproto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

var ErrShort = errors.New("short buffer")

type Buf struct {
	b   []byte
	pos int
}

func New() *Buf                 { return &Buf{} }
func NewReader(b []byte) *Buf   { return &Buf{b: b} }
func (w *Buf) Bytes() []byte    { return w.b }
func (w *Buf) Len() int         { return len(w.b) - w.pos }
func (w *Buf) Remaining() []byte { return w.b[w.pos:] }

func (w *Buf) WriteVarInt(v int32) {
	uv := uint32(v)
	for {
		b := byte(uv & 0x7f)
		uv >>= 7
		if uv != 0 {
			b |= 0x80
		}
		w.b = append(w.b, b)
		if uv == 0 {
			return
		}
	}
}

func (w *Buf) WriteU8(v byte)      { w.b = append(w.b, v) }
func (w *Buf) WriteBool(v bool)    { if v { w.b = append(w.b, 1) } else { w.b = append(w.b, 0) } }
func (w *Buf) WriteI16(v int16)    { w.b = binary.BigEndian.AppendUint16(w.b, uint16(v)) }
func (w *Buf) WriteU16(v uint16)   { w.b = binary.BigEndian.AppendUint16(w.b, v) }
func (w *Buf) WriteI32(v int32)    { w.b = binary.BigEndian.AppendUint32(w.b, uint32(v)) }
func (w *Buf) WriteI64(v int64)    { w.b = binary.BigEndian.AppendUint64(w.b, uint64(v)) }
func (w *Buf) WriteF32(v float32)  { w.b = binary.BigEndian.AppendUint32(w.b, math.Float32bits(v)) }
func (w *Buf) WriteF64(v float64)  { w.b = binary.BigEndian.AppendUint64(w.b, math.Float64bits(v)) }
func (w *Buf) WriteRaw(p []byte)   { w.b = append(w.b, p...) }

func (w *Buf) WriteString(s string) {
	w.WriteVarInt(int32(len(s)))
	w.b = append(w.b, s...)
}

func (w *Buf) WritePosition(x, y, z int) {
	v := (uint64(x&0x3FFFFFF) << 38) | (uint64(y&0xFFF) << 26) | uint64(z&0x3FFFFFF)
	w.WriteI64(int64(v))
}

func (w *Buf) ReadVarInt() (int32, error) {
	var res uint32
	for i := 0; i < 5; i++ {
		if w.pos >= len(w.b) {
			return 0, ErrShort
		}
		b := w.b[w.pos]
		w.pos++
		res |= uint32(b&0x7f) << (7 * i)
		if b&0x80 == 0 {
			return int32(res), nil
		}
	}
	return 0, errors.New("varint too long")
}

func (w *Buf) ReadU8() (byte, error) {
	if w.pos >= len(w.b) {
		return 0, ErrShort
	}
	v := w.b[w.pos]
	w.pos++
	return v, nil
}

func (w *Buf) ReadI8() (int8, error)   { v, e := w.ReadU8(); return int8(v), e }
func (w *Buf) ReadBool() (bool, error) { v, e := w.ReadU8(); return v != 0, e }

func (w *Buf) need(n int) error {
	if w.pos+n > len(w.b) {
		return ErrShort
	}
	return nil
}

func (w *Buf) ReadU16() (uint16, error) {
	if err := w.need(2); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint16(w.b[w.pos:])
	w.pos += 2
	return v, nil
}

func (w *Buf) ReadI16() (int16, error) { v, e := w.ReadU16(); return int16(v), e }

func (w *Buf) ReadI32() (int32, error) {
	if err := w.need(4); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint32(w.b[w.pos:])
	w.pos += 4
	return int32(v), nil
}

func (w *Buf) ReadI64() (int64, error) {
	if err := w.need(8); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint64(w.b[w.pos:])
	w.pos += 8
	return int64(v), nil
}

func (w *Buf) ReadF32() (float32, error) {
	v, e := w.ReadI32()
	return math.Float32frombits(uint32(v)), e
}

func (w *Buf) ReadF64() (float64, error) {
	v, e := w.ReadI64()
	return math.Float64frombits(uint64(v)), e
}

func (w *Buf) ReadString() (string, error) {
	n, err := w.ReadVarInt()
	if err != nil {
		return "", err
	}
	if n < 0 || w.pos+int(n) > len(w.b) {
		return "", ErrShort
	}
	s := string(w.b[w.pos : w.pos+int(n)])
	w.pos += int(n)
	return s, nil
}

func (w *Buf) ReadPosition() (x, y, z int, err error) {
	v, err := w.ReadI64()
	if err != nil {
		return 0, 0, 0, err
	}
	uv := uint64(v)
	x = int(uv >> 38)
	y = int((uv >> 26) & 0xFFF)
	z = int(uv & 0x3FFFFFF)
	if x >= 1<<25 {
		x -= 1 << 26
	}
	if y >= 1<<11 {
		y -= 1 << 12
	}
	if z >= 1<<25 {
		z -= 1 << 26
	}
	return x, y, z, nil
}

func Frame(id int32, body []byte) []byte {
	inner := New()
	inner.WriteVarInt(id)
	inner.WriteRaw(body)
	out := New()
	out.WriteVarInt(int32(inner.Len()))
	out.WriteRaw(inner.Bytes())
	return out.Bytes()
}

type Packet struct {
	ID   int32
	Body *Buf
}

func ReadPackets(buf []byte) ([]Packet, []byte) {
	var out []Packet
	i := 0
	for {
		r := &Buf{b: buf, pos: i}
		length, err := r.ReadVarInt()
		if err != nil {
			break
		}
		start := r.pos
		if start+int(length) > len(buf) {
			break
		}
		body := &Buf{b: buf[start : start+int(length)]}
		id, err := body.ReadVarInt()
		if err != nil {
			break
		}
		out = append(out, Packet{ID: id, Body: body})
		i = start + int(length)
	}
	return out, buf[i:]
}

func WriteFramed(w io.Writer, id int32, body []byte) error {
	_, err := w.Write(Frame(id, body))
	return err
}

func Concat(parts ...[]byte) []byte { return bytes.Join(parts, nil) }
