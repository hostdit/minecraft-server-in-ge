package kml

import (
	"fmt"
	"strings"
)

type Block struct {
	Lat, Lng float64
	ABGR     string
	Label    string
}

type WonderView struct {
	Name     string
	Lat, Lng float64
	Honorary bool
	Current  bool
}

type Snapshot struct {
	Blocks    []Block
	Wonders   []WonderView
	PlayerLat float64
	PlayerLng float64
	HasPlayer bool
	Placed    int
	FlyTo     bool
	FlyLat    float64
	FlyLng    float64
}

func esc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func Bootstrap(url string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Document>
    <name>Minecraft on Earth</name>
    <NetworkLink>
      <name>Live world</name>
      <flyToView>1</flyToView>
      <Link>
        <href>` + esc(url) + `</href>
        <refreshMode>onInterval</refreshMode>
        <refreshInterval>2</refreshInterval>
      </Link>
    </NetworkLink>
  </Document>
</kml>`
}

func Live(s Snapshot) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<kml xmlns="http://www.opengis.net/kml/2.2" xmlns:gx="http://www.google.com/kml/ext/2.2">` + "\n")
	if s.FlyTo {
		fmt.Fprintf(&b,
			"<NetworkLinkControl><LookAt><longitude>%.6f</longitude><latitude>%.6f</latitude><altitude>0</altitude><range>600</range><tilt>65</tilt><heading>0</heading><altitudeMode>relativeToGround</altitudeMode></LookAt></NetworkLinkControl>\n",
			s.FlyLng, s.FlyLat)
	}
	b.WriteString("<Document>\n")
	b.WriteString("<name>Minecraft world (" + fmt.Sprintf("%d blocks placed", s.Placed) + ")</name>\n")

	b.WriteString(`<Style id="block"><IconStyle><scale>0.5</scale><Icon><href>http://maps.google.com/mapfiles/kml/shapes/placemark_square.png</href></Icon></IconStyle><LabelStyle><scale>0</scale></LabelStyle></Style>` + "\n")
	b.WriteString(`<Style id="wonder"><IconStyle><scale>1.3</scale><Icon><href>http://maps.google.com/mapfiles/kml/shapes/star.png</href></Icon></IconStyle></Style>` + "\n")
	b.WriteString(`<Style id="player"><IconStyle><scale>1.4</scale><Icon><href>http://maps.google.com/mapfiles/kml/shapes/man.png</href></Icon></IconStyle></Style>` + "\n")

	b.WriteString("<Folder><name>Wonders</name>\n")
	for _, w := range s.Wonders {
		tag := ""
		if w.Honorary {
			tag = " (honorary)"
		}
		if w.Current {
			tag += " - you are here"
		}
		fmt.Fprintf(&b,
			"<Placemark><name>%s%s</name><styleUrl>#wonder</styleUrl><Point><coordinates>%.6f,%.6f,0</coordinates></Point></Placemark>\n",
			esc(w.Name), tag, w.Lng, w.Lat)
	}
	b.WriteString("</Folder>\n")

	b.WriteString("<Folder><name>Placed blocks</name>\n")
	for _, bl := range s.Blocks {
		fmt.Fprintf(&b,
			"<Placemark><name>%s</name><styleUrl>#block</styleUrl><Style><IconStyle><color>%s</color><scale>0.5</scale><Icon><href>http://maps.google.com/mapfiles/kml/shapes/placemark_square.png</href></Icon></IconStyle><LabelStyle><scale>0</scale></LabelStyle></Style><Point><coordinates>%.6f,%.6f,0</coordinates></Point></Placemark>\n",
			esc(bl.Label), bl.ABGR, bl.Lng, bl.Lat)
	}
	b.WriteString("</Folder>\n")

	if s.HasPlayer {
		fmt.Fprintf(&b,
			"<Placemark><name>Player</name><styleUrl>#player</styleUrl><Point><coordinates>%.6f,%.6f,0</coordinates></Point></Placemark>\n",
			s.PlayerLng, s.PlayerLat)
	}

	b.WriteString("</Document>\n</kml>\n")
	return b.String()
}

func Tour(ws []WonderView) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<kml xmlns="http://www.opengis.net/kml/2.2" xmlns:gx="http://www.google.com/kml/ext/2.2">` + "\n")
	b.WriteString("<gx:Tour><name>Seven Wonders fly-through</name><gx:Playlist>\n")
	for i, w := range ws {
		dur := 4.0
		if i == 0 {
			dur = 2.0
		}
		fmt.Fprintf(&b,
			"<gx:FlyTo><gx:duration>%.1f</gx:duration><gx:flyToMode>smooth</gx:flyToMode><LookAt><longitude>%.6f</longitude><latitude>%.6f</latitude><altitude>0</altitude><range>1200</range><tilt>60</tilt><heading>0</heading><altitudeMode>relativeToGround</altitudeMode></LookAt></gx:FlyTo>\n",
			dur, w.Lng, w.Lat)
		b.WriteString("<gx:Wait><gx:duration>2.0</gx:duration></gx:Wait>\n")
	}
	b.WriteString("</gx:Playlist></gx:Tour>\n</kml>\n")
	return b.String()
}
