package mcproto

import "strings"

func StatusResponse(json string) []byte {
	b := New()
	b.WriteString(json)
	return Frame(0x00, b.Bytes())
}

func Pong(payload int64) []byte {
	b := New()
	b.WriteI64(payload)
	return Frame(0x01, b.Bytes())
}

func LoginSuccess(uuid, name string) []byte {
	b := New()
	b.WriteString(uuid)
	b.WriteString(name)
	return Frame(0x02, b.Bytes())
}

func JoinGame(eid int32) []byte {
	b := New()
	b.WriteI32(eid)
	b.WriteU8(1)
	b.WriteU8(0)
	b.WriteU8(1)
	b.WriteU8(1)
	b.WriteString("flat")
	b.WriteBool(false)
	return Frame(0x01, b.Bytes())
}

func SpawnPosition(x, y, z int) []byte {
	b := New()
	b.WritePosition(x, y, z)
	return Frame(0x05, b.Bytes())
}

func PlayerAbilities() []byte {
	b := New()
	b.WriteU8(0x0D)
	b.WriteF32(0.05)
	b.WriteF32(0.1)
	return Frame(0x39, b.Bytes())
}

func PositionLook(x, y, z float64, yaw, pitch float32) []byte {
	b := New()
	b.WriteF64(x)
	b.WriteF64(y)
	b.WriteF64(z)
	b.WriteF32(yaw)
	b.WriteF32(pitch)
	b.WriteU8(0)
	return Frame(0x08, b.Bytes())
}

func KeepAlive(id int32) []byte {
	b := New()
	b.WriteVarInt(id)
	return Frame(0x00, b.Bytes())
}

func BlockChange(x, y, z int, id, meta int) []byte {
	b := New()
	b.WritePosition(x, y, z)
	b.WriteVarInt(int32(id<<4 | (meta & 0xF)))
	return Frame(0x23, b.Bytes())
}

func chatJSON(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "\"", "\\\"")
	return `{"text":"` + text + `"}`
}

func Chat(text string) []byte {
	b := New()
	b.WriteString(chatJSON(text))
	b.WriteU8(0)
	return Frame(0x02, b.Bytes())
}

func Disconnect(text string) []byte {
	b := New()
	b.WriteString(chatJSON(text))
	return Frame(0x40, b.Bytes())
}

func Chunk(cx, cz int32, layers [][2]int) []byte {
	const section = 16 * 16 * 16
	blocks := make([]byte, section*2)
	for y := 0; y < 16; y++ {
		var v int
		if y < len(layers) {
			v = layers[y][0]<<4 | (layers[y][1] & 0xF)
		}
		lo := byte(v & 0xFF)
		hi := byte((v >> 8) & 0xFF)
		base := y * 256 * 2
		for i := 0; i < 256; i++ {
			blocks[base+i*2] = lo
			blocks[base+i*2+1] = hi
		}
	}
	data := New()
	data.WriteRaw(blocks)
	data.WriteRaw(make([]byte, 2048))
	skylight := make([]byte, 2048)
	for i := range skylight {
		skylight[i] = 0xFF
	}
	data.WriteRaw(skylight)
	data.WriteRaw(make([]byte, 256))

	b := New()
	b.WriteI32(cx)
	b.WriteI32(cz)
	b.WriteBool(true)
	b.WriteU16(0x0001)
	b.WriteVarInt(int32(data.Len()))
	b.WriteRaw(data.Bytes())
	return Frame(0x21, b.Bytes())
}
