package mcserver

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"gemc/internal/geo"
	"gemc/internal/geomap"
	"gemc/internal/kml"
	"gemc/internal/mcproto"
	"gemc/internal/wonders"
)

const (
	spawnX, spawnZ = 8, 8
	groundY        = 4
	metresPerBlock = 1.0
)

var floor = [][2]int{{7, 0}, {3, 0}, {3, 0}, {2, 0}}

type earthBlock struct {
	lat, lng float64
	abgr     string
	label    string
}

type sceneBlock struct {
	id, meta int
	byPlayer bool
}

type World struct {
	mu       sync.Mutex
	favicon  string
	anchor   geomap.Anchor
	wonder   string
	scene    map[[3]int]sceneBlock
	earth    map[string]earthBlock
	placed   int
	playerX  float64
	playerY  float64
	playerZ  float64
	hasPlay  bool
	flyTo    bool
	flyLat   float64
	flyLng   float64
}

func NewWorld() *World {
	w := &World{scene: map[[3]int]sceneBlock{}, earth: map[string]earthBlock{}}
	w.setWonder("giza")
	w.loadFavicon("favicon.png")
	return w
}

func (w *World) loadFavicon(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if len(data) < 24 || string(data[1:4]) != "PNG" {
		log.Printf("favicon %s is not a PNG, ignoring", path)
		return
	}
	wpx := int(data[16])<<24 | int(data[17])<<16 | int(data[18])<<8 | int(data[19])
	hpx := int(data[20])<<24 | int(data[21])<<16 | int(data[22])<<8 | int(data[23])
	if wpx != 64 || hpx != 64 {
		log.Printf("favicon %s is %dx%d, must be exactly 64x64, ignoring", path, wpx, hpx)
		return
	}
	w.favicon = "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	log.Printf("favicon loaded from %s (%d bytes)", path, len(data))
}

func (w *World) setWonder(key string) {
	won, ok := wonders.ByKey(key)
	if !ok {
		return
	}
	w.wonder = key
	w.flyTo, w.flyLat, w.flyLng = true, won.Lat, won.Lng
	w.anchor = geomap.Anchor{Lat: won.Lat, Lng: won.Lng, BlockX: spawnX, BlockZ: spawnZ, MetresPerBlock: metresPerBlock}
	w.scene = map[[3]int]sceneBlock{}
	for _, bl := range won.Footprint() {
		if bl.ID == 0 {
			continue
		}
		w.scene[[3]int{bl.X + spawnX, groundY + 1 + bl.Y, bl.Z + spawnZ}] = sceneBlock{bl.ID, bl.Meta, false}
	}
}

func abgrFor(id, meta int) string {
	switch id {
	case 1, 98:
		return "ff888888"
	case 2:
		return "ff3cb43c"
	case 4:
		return "ff5a5a5a"
	case 24:
		return "ff8ecbd8"
	case 41:
		return "ff3ce6fa"
	case 155:
		return "fff5f5f5"
	case 20:
		return "ffffd8b4"
	case 35:
		w := [][3]int{{234, 234, 234}, {53, 125, 235}, {200, 73, 190}, {211, 138, 105}, {40, 182, 195}, {50, 180, 60}, {158, 130, 216}, {63, 63, 63}, {165, 165, 155}, {140, 120, 40}, {180, 60, 125}, {145, 55, 45}, {30, 55, 80}, {25, 75, 55}, {35, 40, 165}, {22, 22, 25}}
		c := w[meta%16]
		return fmt.Sprintf("ff%02x%02x%02x", c[2], c[1], c[0])
	default:
		return "ffb400b4"
	}
}

func (w *World) TakeSnapshot() kml.Snapshot {
	s := w.Snapshot()
	w.mu.Lock()
	w.flyTo = false
	w.mu.Unlock()
	return s
}

func (w *World) Snapshot() kml.Snapshot {
	w.mu.Lock()
	defer w.mu.Unlock()
	s := kml.Snapshot{Placed: w.placed, HasPlayer: w.hasPlay, FlyTo: w.flyTo, FlyLat: w.flyLat, FlyLng: w.flyLng}
	for _, e := range w.earth {
		s.Blocks = append(s.Blocks, kml.Block{Lat: e.lat, Lng: e.lng, ABGR: e.abgr, Label: e.label})
	}
	for _, won := range wonders.List {
		s.Wonders = append(s.Wonders, kml.WonderView{Name: won.Name, Lat: won.Lat, Lng: won.Lng, Honorary: won.Honorary, Current: won.Key == w.wonder})
	}
	if w.hasPlay {
		s.PlayerLat, s.PlayerLng = w.anchor.ToGeo(int(w.playerX), int(w.playerZ))
	}
	return s
}

type conn struct {
	c     net.Conn
	w     *World
	buf   []byte
	state int
	name  string
	out   chan []byte
}

func (cn *conn) send(p []byte) {
	select {
	case cn.out <- p:
	default:
	}
}

func Run(mcAddr, httpAddr, publicURL string, w *World) error {
	http.HandleFunc("/live.kml", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/vnd.google-earth.kml+xml")
		fmt.Fprint(rw, kml.Live(w.TakeSnapshot()))
	})
	http.HandleFunc("/earth.kml", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/vnd.google-earth.kml+xml")
		rw.Header().Set("Content-Disposition", `attachment; filename="minecraft-on-earth.kml"`)
		fmt.Fprint(rw, kml.Bootstrap(publicURL+"/live.kml"))
	})
	http.HandleFunc("/tour.kml", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/vnd.google-earth.kml+xml")
		rw.Header().Set("Content-Disposition", `attachment; filename="wonders-tour.kml"`)
		s := w.Snapshot()
		fmt.Fprint(rw, kml.Tour(s.Wonders))
	})
	go func() {
		log.Printf("KML endpoint on http://%s  (open %s/earth.kml in Google Earth)", httpAddr, publicURL)
		if err := http.ListenAndServe(httpAddr, nil); err != nil {
			log.Fatalf("http: %v", err)
		}
	}()

	ln, err := net.Listen("tcp", mcAddr)
	if err != nil {
		return err
	}
	log.Printf("Minecraft server listening on %s", mcAddr)
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		cn := &conn{c: c, w: w, state: 0, out: make(chan []byte, 4096)}
		go cn.writeLoop()
		go cn.readLoop()
	}
}

func (cn *conn) writeLoop() {
	broken := false
	for p := range cn.out {
		if broken {
			continue
		}
		if _, err := cn.c.Write(p); err != nil {
			broken = true
		}
	}
	cn.c.Close()
}

func (cn *conn) readLoop() {
	defer func() {
		close(cn.out)
		if cn.state == 3 {
			cn.w.mu.Lock()
			cn.w.hasPlay = false
			cn.w.mu.Unlock()
			log.Printf("%s disconnected", cn.name)
		}
	}()
	tmp := make([]byte, 32768)
	for {
		n, err := cn.c.Read(tmp)
		if err != nil {
			return
		}
		cn.buf = append(cn.buf, tmp[:n]...)
		pkts, rest := mcproto.ReadPackets(cn.buf)
		cn.buf = rest
		for _, p := range pkts {
			if !cn.handle(p) {
				return
			}
		}
	}
}

func (cn *conn) handle(p mcproto.Packet) bool {
	switch cn.state {
	case 0:
		_, _ = p.Body.ReadVarInt()
		_, _ = p.Body.ReadString()
		_, _ = p.Body.ReadU16()
		next, _ := p.Body.ReadVarInt()
		cn.state = int(next)
	case 1:
		switch p.ID {
		case 0x00:
			cn.send(mcproto.StatusResponse(cn.statusJSON()))
		case 0x01:
			payload, _ := p.Body.ReadI64()
			cn.send(mcproto.Pong(payload))
			return false
		}
	case 2:
		if p.ID == 0x00 {
			cn.name, _ = p.Body.ReadString()
			cn.join()
		}
	case 3:
		cn.play(p)
	}
	return true
}

func (cn *conn) statusJSON() string {
	online := 0
	cn.w.mu.Lock()
	if cn.w.hasPlay {
		online = 1
	}
	cn.w.mu.Unlock()
	cn.w.mu.Lock()
	fav := cn.w.favicon
	cn.w.mu.Unlock()
	favField := ""
	if fav != "" {
		favField = fmt.Sprintf(`,"favicon":"%s"`, fav)
	}
	return fmt.Sprintf(`{"version":{"name":"Earth","protocol":47},"players":{"max":1,"online":%d},"description":{"text":"A Minecraft world on planet earth"}%s}`, online, favField)
}

func (cn *conn) join() {
	cn.w.mu.Lock()
	cn.w.hasPlay = true
	cn.w.playerX, cn.w.playerY, cn.w.playerZ = spawnX, groundY+1, spawnZ
	cn.state = 3
	cn.w.mu.Unlock()

	cn.send(mcproto.LoginSuccess("00000000-0000-3000-8000-000000000001", cn.name))
	cn.send(mcproto.JoinGame(1))
	cn.send(mcproto.SpawnPosition(spawnX, groundY, spawnZ))
	cn.send(mcproto.PlayerAbilities())
	for cx := int32(-2); cx <= 2; cx++ {
		for cz := int32(-2); cz <= 2; cz++ {
			cn.send(mcproto.Chunk(cx, cz, floor))
		}
	}
	cn.sendScene()
	cn.send(mcproto.PositionLook(spawnX+0.5, groundY+1, spawnZ+0.5, 0, 0))
	won, _ := wonders.ByKey(cn.w.wonder)
	cn.send(mcproto.Chat("You're at " + won.Name + ". Every block you place lands where you stand on the real planet."))
	cn.send(mcproto.Chat("Type /list to see the wonders, /warp <name> to travel."))
	log.Printf("%s joined at %s", cn.name, won.Name)
	go cn.keepAlive()
}

func (cn *conn) sendScene() {
	cn.w.mu.Lock()
	type cell struct {
		pos      [3]int
		id, meta int
	}
	var cells []cell
	for k, v := range cn.w.scene {
		cells = append(cells, cell{k, v.id, v.meta})
	}
	cn.w.mu.Unlock()
	for _, c := range cells {
		cn.send(mcproto.BlockChange(c.pos[0], c.pos[1], c.pos[2], c.id, c.meta))
	}
}

func (cn *conn) keepAlive() {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	id := int32(0)
	for range t.C {
		id++
		select {
		case cn.out <- mcproto.KeepAlive(id):
		default:
			return
		}
	}
}

func (cn *conn) play(p mcproto.Packet) {
	switch p.ID {
	case 0x04:
		x, _ := p.Body.ReadF64()
		y, _ := p.Body.ReadF64()
		z, _ := p.Body.ReadF64()
		cn.setPos(x, y, z)
	case 0x06:
		x, _ := p.Body.ReadF64()
		y, _ := p.Body.ReadF64()
		z, _ := p.Body.ReadF64()
		cn.setPos(x, y, z)
	case 0x07:
		status, _ := p.Body.ReadI8()
		bx, by, bz, _ := p.Body.ReadPosition()
		if status == 0 {
			cn.dig(bx, by, bz)
		}
	case 0x08:
		bx, by, bz, _ := p.Body.ReadPosition()
		face, _ := p.Body.ReadI8()
		id, _ := p.Body.ReadI16()
		if id <= 0 || face < 0 {
			return
		}
		meta := 0
		if _, err := p.Body.ReadU8(); err == nil {
			d, _ := p.Body.ReadI16()
			meta = int(d)
		}
		switch face {
		case 0:
			by--
		case 1:
			by++
		case 2:
			bz--
		case 3:
			bz++
		case 4:
			bx--
		case 5:
			bx++
		}
		cn.place(bx, by, bz, int(id), meta)
	case 0x01:
		msg, _ := p.Body.ReadString()
		cn.chat(msg)
	}
}

func (cn *conn) setPos(x, y, z float64) {
	cn.w.mu.Lock()
	cn.w.playerX, cn.w.playerY, cn.w.playerZ = x, y, z
	cn.w.mu.Unlock()
}

func (cn *conn) place(x, y, z, id, meta int) {
	cn.w.mu.Lock()
	cell := [3]int{x, y, z}
	cn.w.scene[cell] = sceneBlock{id, meta, true}
	lat, lng := cn.w.anchor.ToGeo(x, z)
	key := fmt.Sprintf("%s|%d,%d,%d", cn.w.wonder, x, y, z)
	cn.w.placed++
	anchor := cn.w.anchor
	cn.w.mu.Unlock()

	cn.send(mcproto.BlockChange(x, y, z, id, meta))
	_ = anchor
	label := geo.Describe(lat, lng)
	cn.w.mu.Lock()
	cn.w.earth[key] = earthBlock{lat, lng, abgrFor(id, meta), fmt.Sprintf("Block in %s", label)}
	cn.w.mu.Unlock()
	cn.send(mcproto.Chat(fmt.Sprintf("Placed in %s", label)))
}

func (cn *conn) dig(x, y, z int) {
	cn.w.mu.Lock()
	cell := [3]int{x, y, z}
	delete(cn.w.scene, cell)
	key := fmt.Sprintf("%s|%d,%d,%d", cn.w.wonder, x, y, z)
	delete(cn.w.earth, key)
	cn.w.mu.Unlock()
	cn.send(mcproto.BlockChange(x, y, z, 0, 0))
}

func (cn *conn) chat(msg string) {
	if !strings.HasPrefix(msg, "/") {
		cn.send(mcproto.Chat("<" + cn.name + "> " + msg))
		return
	}
	fields := strings.Fields(msg)
	switch fields[0] {
	case "/list":
		var names []string
		for _, w := range wonders.List {
			names = append(names, w.Key)
		}
		cn.send(mcproto.Chat("Wonders: " + strings.Join(names, ", ")))
	case "/warp":
		if len(fields) < 2 {
			cn.send(mcproto.Chat("Usage: /warp <name>"))
			return
		}
		won, ok := wonders.ByKey(fields[1])
		if !ok {
			cn.send(mcproto.Chat("No wonder called " + fields[1]))
			return
		}
		cn.warp(won)
	default:
		cn.send(mcproto.Chat("Unknown command"))
	}
}

func (cn *conn) warp(won wonders.Wonder) {
	cn.w.mu.Lock()
	old := make([][3]int, 0, len(cn.w.scene))
	for k := range cn.w.scene {
		old = append(old, k)
	}
	cn.w.setWonder(won.Key)
	cn.w.playerX, cn.w.playerY, cn.w.playerZ = spawnX, groundY+1, spawnZ
	cn.w.mu.Unlock()

	for _, k := range old {
		cn.send(mcproto.BlockChange(k[0], k[1], k[2], 0, 0))
	}
	cn.sendScene()
	cn.send(mcproto.PositionLook(spawnX+0.5, groundY+1, spawnZ+0.5, 0, 0))
	cn.send(mcproto.Chat("Warped to " + won.Name + "."))
	log.Printf("%s warped to %s", cn.name, won.Name)
}
