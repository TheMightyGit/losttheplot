package cartridge

import (
	"embed"
	"fmt"
	"image"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/TheMightyGit/marv/marvlib"
	"github.com/TheMightyGit/marv/marvtypes"
)

//go:embed "resources/*"
var Resources embed.FS

const (
	MODE_TITLE_SCREEN_SETUP = iota
	MODE_TITLE_SCREEN
	MODE_GAME_SETUP
	MODE_GAME
)

const (
	GfxBankFont = iota
	GfxBankGraveyard
	GfxBankPeople
	GfxBankTitles
)
const (
	MapBankText = iota
	MapBankGraveyard
	MapBankPeople
)
const (
	SpriteGraveyardUnderlay = iota
	SpriteGraveyard
	SpriteGraveyardOverlay
	SpriteVisitor1
	SpriteVisitor2
	SpriteVisitor3
	SpriteVisitor4
	SpriteVisitor5
	SpritePlayer
	SpriteText
	SpriteMousePointer
)

var (
	mode            int
	fullScreenRect  = image.Rectangle{Max: image.Point{X: 320, Y: 200}}
	titleScreenRect = image.Rectangle{Min: image.Point{32 + 8, -40}, Max: image.Point{X: 256, Y: 256}}
)

type Camera struct {
	X float64
	Y float64

	TargetX float64
	TargetY float64
}

func (c *Camera) GetAsPoint() image.Point {
	return image.Point{X: int(c.X), Y: int(c.Y)}
}

func (c *Camera) Shake() {
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(16 * time.Millisecond)
			c.X += (rand.Float64() - 0.5) * 5
			c.Y += (rand.Float64() - 0.5) * 5
		}
	}()
}

func (c *Camera) Update() {
	if c.X+160 < c.TargetX {
		if c.X < float64((graveyard.w*4*8)-320) {
			c.X += 1
		}
	}
	if c.X+160 > c.TargetX {
		if c.X > 1 {
			c.X -= 1
		}
	}
	if c.Y+100 < c.TargetY {
		if c.Y < float64((graveyard.h*4*8)-200) {
			c.Y += 1
		}
	}
	if c.Y+100 > c.TargetY {
		if c.Y > 1 {
			c.Y -= 1
		}
	}
}

type Player struct {
	X     float64
	Y     float64
	Speed float64
	w     int
	h     int
	tx    float64
	ty    float64

	animFrame int

	flagTalkedToVisitors bool
	flagDoneSomeDigging  bool
}

func NewPlayer() *Player {
	p := &Player{
		Speed: 2,
		X:     128,
		Y:     128,
		w:     32,
		h:     64,
	}

	marvlib.API.SpritesGet(SpritePlayer).Show(GfxBankPeople, areaPeople)

	return p
}

func (p *Player) Update() {

	if text.plotInput {
		if text.selectedHoriz != "" && text.selectedVertical != "" {
			// fmt.Println("PLOT INPUT!!")
			text.PlayerSay("") // NOTE: blank needed to stop plot input.
			text.PlayerSay("They're 100% in " + text.selectedVertical + text.selectedHoriz + ".\nNo doubt. I'm almost certain\nthat I'm probably right.")

			// how do we know which visitor to check with?
			visitors.currentVisitor.ChoosePlot(text.selectedVertical, text.selectedHoriz)
		}
	}

	if marvlib.API.InputMousePressed() {
		clickPoint := marvlib.API.InputMousePos()

		if text.plotInput && text.Hitbox.IsHitNoCameraOffset(clickPoint) {
			// let text component handle the input itself.
		} else if grave := graveyard.GetClickedGrave(clickPoint); grave != nil {
			if !p.flagTalkedToVisitors {
				text.PlayerSay("I should talk to the visitors\nbefore I get digging!")
			} else {
				p.flagDoneSomeDigging = true
				// fmt.Println(grave)
				grave.Dig()
				// move the grave (vertically) into visible space.
				camera.TargetY = camera.Y + float64(clickPoint.Y)
			}
		} else if visitor := visitors.GetClickedVisitor(clickPoint); visitor != nil {
			p.flagTalkedToVisitors = true
			// fmt.Println(visitor)
			// fmt.Println(visitor.Grave)
			text.VisitorSay(
				"Where is my...\n ..."+visitor.Grave.Relation+"?\n\n"+strings.Join(visitor.Grave.Likes, "\n"),
				fmt.Sprintf("Visitor %d", (visitor.SpriteId-SpriteVisitor1)+1),
			)
			camera.TargetX = float64(visitor.Pos.X + 16)
			camera.TargetY = float64(visitor.Pos.Y - 64)
			visitors.currentVisitor = visitor
			if p.flagDoneSomeDigging {
				// show grid ref UI
				text.EnablePlotInput()
			}
		} else {
			// move to here (stay within bounds!)
			p.tx, p.ty = float64(clickPoint.X)+camera.X, float64(clickPoint.Y)+camera.Y-20

			if p.ty < 20 {
				p.ty = 20
			}
			if p.tx < 40 {
				p.tx = 40
			}
			if p.tx > 470 {
				p.tx = 470
			}

			// marvlib.API.ConsolePrintln(p.tx, p.ty)

			camera.TargetX = float64(p.tx)
			camera.TargetY = float64(p.ty)
			text.VisitorSay("", "")
			text.PlayerSay("")
		}

	}

	if p.tx != 0 && p.ty != 0 {

		// are we already wihin p.Speed of the target?
		if math.Abs(p.X-p.tx) > p.Speed {
			// move towards target click
			if p.X < p.tx {
				p.X += p.Speed
				p.animFrame++
			} else if p.X > p.tx {
				p.X -= p.Speed
				p.animFrame++
			}
		}

		if math.Abs(p.Y-p.ty) > p.Speed {
			if p.Y < p.ty {
				p.Y += p.Speed
				p.animFrame++
			} else if p.Y > p.ty {
				p.Y -= p.Speed
				p.animFrame++
			}
		}

		p.animFrame = p.animFrame % 20
	}

	cp := camera.GetAsPoint()
	pos := image.Point{X: int(p.X) - (p.w / 2) - cp.X, Y: int(p.Y) - (p.h / 2) - cp.Y}
	marvlib.API.SpritesGet(SpritePlayer).ChangePos(image.Rectangle{
		Min: pos,
		Max: image.Point{X: p.w, Y: p.h},
	})
	marvlib.API.SpritesGet(SpritePlayer).ChangeViewport(image.Point{X: 32 * (p.animFrame / 10)})

	ySortPos := pos.Y + cp.Y
	if ySortPos < 1 {
		ySortPos = 1
	}
	marvlib.API.SpritesGet(SpritePlayer).SetSortIdx(ySortPos)
	marvlib.API.SpritesSort()
}

type Graveyard struct {
	w      int
	h      int
	graves []*Grave
}

func (g *Graveyard) Setup(numRowsOfGraves int) {
	g.w = 16 // Fixed, or it'll overflow onto the overlay and underlay parts of the map.
	g.h = 4 + (numRowsOfGraves * 4)

	g.clearOverlay()
	g.drawWalls()
	g.drawGraves()
	g.drawGrass()

	// splat mouse pointer somewhere we can use it
	pos := image.Point{}
	for pos.Y = 0; pos.Y < 4; pos.Y++ {
		for pos.X = 0; pos.X < 4; pos.X++ {
			areaMousePointer.Set(pos, uint8(8+pos.X), uint8(20+pos.Y), 0, 0)
		}
	}

	marvlib.API.SpritesGet(SpriteGraveyard).ChangePos(fullScreenRect)
	marvlib.API.SpritesGet(SpriteGraveyard).Show(GfxBankGraveyard, area)

	marvlib.API.SpritesGet(SpriteGraveyardUnderlay).ChangePos(fullScreenRect)
	marvlib.API.SpritesGet(SpriteGraveyardUnderlay).Show(GfxBankGraveyard, areaUnderlay)

	marvlib.API.SpritesGet(SpriteGraveyardOverlay).ChangePos(fullScreenRect)
	marvlib.API.SpritesGet(SpriteGraveyardOverlay).Show(GfxBankGraveyard, areaOverlay)
}

var (
	horizGridRef    = []string{"1", "2", "3", "4", "5"}
	verticalGridRef = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
)

func (g *Graveyard) GetClickedGrave(screenpoint image.Point) *Grave {
	for _, grave := range g.graves {
		if grave.Hitbox.IsHit(screenpoint) {
			return grave
		}
	}
	return nil
}

func (g *Graveyard) drawGrass() {
	for y := 0; y < 256; y += 4 {
		for x := 0; x < 64; x += 4 {
			g.drawCommon(areaUnderlay, x, y, 24, 24, 4, 4)
		}
	}
}

func (g *Graveyard) clearOverlay() {
	areaOverlay.Clear(31, 0)
}

func (g *Graveyard) drawGraves() {
	gX := 0
	gY := 0
	for y := 16; y < ((g.h - 3) * 4); y += 16 {
		for x := 8; x < ((g.w - 1) * 4); x += 11 {
			grave := NewGrave(x, y, horizGridRef[gX], verticalGridRef[gY])
			// fmt.Println(grave)
			g.graves = append(g.graves, grave)

			grave.drawClosedGrave()

			gX++
		}
		gX = 0
		gY++
	}
}

func (g *Graveyard) drawWalls() {
	area.Clear(31, 0)

	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			if y == 0 {
				if x == 0 {
					g.drawTopLeft(x*4, y*4)
				} else if x == (g.w - 1) {
					g.drawTopRight(x*4, y*4)
				} else {
					g.drawTop(x*4, y*4)
				}
			} else if x == 0 {
				if y == (g.h - 1) {
					g.drawBottomLeft(x*4, y*4)
				} else {
					g.drawLeft(x*4, y*4)
				}
			} else if x == (g.w - 1) {
				if y == (g.h - 1) {
					g.drawBottomRight(x*4, y*4)
				} else {
					g.drawRight(x*4, y*4)
				}
			} else if y == (g.h - 1) {
				g.drawBottom(x*4, y*4)
			}
		}
	}
}

func (g *Graveyard) drawCommon(area marvtypes.MapBankArea, dstX int, dstY int, srcX int, srcY int, w int, h int) {
	offset := image.Point{X: dstX, Y: dstY}
	pos := image.Point{}
	for pos.Y = 0; pos.Y < h; pos.Y++ {
		for pos.X = 0; pos.X < w; pos.X++ {
			area.Set(pos.Add(offset), uint8(srcX+pos.X), uint8(srcY+pos.Y), 0, 0)
		}
	}
}

func (g *Graveyard) drawTopLeft(x, y int) {
	g.drawCommon(area, x, y, 20, 20, 4, 4)
}

func (g *Graveyard) drawTop(x, y int) {
	g.drawCommon(area, x, y, 24, 20, 4, 4)
}

func (g *Graveyard) drawTopRight(x, y int) {
	g.drawCommon(area, x, y, 28, 20, 4, 4)
}

func (g *Graveyard) drawLeft(x, y int) {
	g.drawCommon(area, x, y, 20, 24, 4, 4)
}

func (g *Graveyard) drawRight(x, y int) {
	g.drawCommon(area, x, y, 28, 24, 4, 4)
}

func (g *Graveyard) drawBottomLeft(x, y int) {
	g.drawCommon(area, x, y, 20, 28, 4, 4)
}

func (g *Graveyard) drawBottom(x, y int) {
	g.drawCommon(area, x, y, 24, 28, 4, 4)
}

func (g *Graveyard) drawBottomRight(x, y int) {
	g.drawCommon(area, x, y, 28, 28, 4, 4)
}

func (g *Graveyard) Update() {
	cp := camera.GetAsPoint()
	marvlib.API.SpritesGet(SpriteGraveyardUnderlay).ChangeViewport(cp)
	marvlib.API.SpritesGet(SpriteGraveyard).ChangeViewport(cp)
	marvlib.API.SpritesGet(SpriteGraveyardOverlay).ChangeViewport(cp)
}

type Text struct {
	Hitbox           Hitbox
	plotInput        bool
	selectedHoriz    string
	selectedVertical string
}

func NewText() *Text {
	t := &Text{
		Hitbox: Hitbox{
			Rectangle: image.Rectangle{
				Min: image.Point{X: 0, Y: 152},
				Max: image.Point{X: 320, Y: 200},
			},
		},
	}

	areaText.Clear(0, 0)

	t.Clear(0, 0, 54, 25)
	// t.PlayerSay("DIGGING IT")

	marvlib.API.SpritesGet(SpriteText).ChangePos(fullScreenRect)
	marvlib.API.SpritesGet(SpriteText).Show(GfxBankFont, areaText)
	marvlib.API.SpritesGet(SpriteText).SetSortIdx(88888)

	marvlib.API.SpritesGet(SpriteMousePointer).Show(GfxBankGraveyard, areaMousePointer)
	marvlib.API.SpritesGet(SpriteMousePointer).SetSortIdx(99999)

	return t
}

func (t *Text) Clear(sx, sy, w, h int) {
	pos := image.Point{}
	for pos.Y = sy; pos.Y < sy+h; pos.Y++ {
		for pos.X = sx; pos.X < sx+w; pos.X++ {
			areaText.Set(pos, 17, 0, 16, 16)
		}
	}
}

func (t *Text) EnablePlotInput() {
	t.plotInput = true
	t.selectedHoriz = ""
	t.selectedVertical = ""
}

func (t *Text) PlayerSay(txt string) {
	if txt == "" {
		t.Clear(0, 20, 54, 5)
		t.plotInput = false
	} else {
		t.SpeechBox(13, 20, 40, 5, 12)
		areaText.StringToMap(image.Point{X: 1 + 13, Y: 21}, 10, 7, txt)
		areaText.StringToMap(image.Point{X: 45, Y: 20}, 7, 12, " You ")
	}
}

func (t *Text) VisitorSay(txt string, visitorName string) {
	if txt == "" {
		t.Clear(0, 0, 40, 8)
	} else {
		t.SpeechBox(0, 0, 40, 8, 12)
		areaText.StringToMap(image.Point{X: 1, Y: 1}, 10, 7, txt)
		areaText.StringToMap(image.Point{X: 36 - len(visitorName), Y: 0}, 7, 12, " "+visitorName+" ")
		if strings.HasPrefix(visitorName, "Visitor") {
			areaText.Set(image.Point{X: 23, Y: 7}, 18, 2, 12, 16)
		}
	}
}

func (t *Text) SpeechBox(sx, sy, w, h int, col uint8) {
	w += sx
	h += sy
	pos := image.Point{}
	for pos.Y = sy; pos.Y < h; pos.Y++ {
		for pos.X = sx; pos.X < w; pos.X++ {
			if pos.Y == sy {
				if pos.X == sx {
					areaText.Set(pos, 17, 0, col, 16)
				} else if pos.X == w-1 {
					areaText.Set(pos, 20, 0, col, 16)
				} else {
					areaText.Set(pos, 18, 0, col, 16)
				}
			} else if pos.Y == h-1 {
				if pos.X == sx {
					areaText.Set(pos, 17, 2, col, 16)
				} else if pos.X == w-1 {
					areaText.Set(pos, 20, 2, col, 16)
				} else {
					areaText.Set(pos, 19, 2, col, 16)
				}
			} else if pos.X == sx {
				areaText.Set(pos, 17, 1, col, 16)
			} else if pos.X == w-1 {
				areaText.Set(pos, 20, 1, col, 16)
			} else {
				areaText.Set(pos, 16, 0, col, 7)
			}
		}
	}
}

func (t *Text) updatePlotUI() {
	text.PlayerSay("They're in plot...") //  " + strings.Join(verticalGridRef, "  ") + "\n  " + strings.Join(horizGridRef, "  "))

	mousePos := marvlib.API.InputMousePos()
	mousePos.X -= 6 * 11

	for i, v := range verticalGridRef {
		if i < lvlNum {
			r := image.Rectangle{
				Min: image.Point{X: 13 + (3+(i*3))*6, Y: 22 * 8},
				Max: image.Point{X: 13 + ((3 + (i * 3)) * 6) + 6, Y: (22 * 8) + 8},
			}
			if mousePos.In(r) || t.selectedVertical == v {
				areaText.StringToMap(image.Point{X: 13 + 3 + (i * 3), Y: 22}, 3, 14, v)
				if marvlib.API.InputMousePressed() {
					t.selectedVertical = v
				}
			} else {
				areaText.StringToMap(image.Point{X: 13 + 3 + (i * 3), Y: 22}, 7, 10, v)
			}
		}
	}
	for i, h := range horizGridRef {
		r := image.Rectangle{
			Min: image.Point{X: 13 + (3+(i*3))*6, Y: 23 * 8},
			Max: image.Point{X: 13 + ((3 + (i * 3)) * 6) + 6, Y: (23 * 8) + 8},
		}
		if mousePos.In(r) || t.selectedHoriz == h {
			areaText.StringToMap(image.Point{X: 13 + 3 + (i * 3), Y: 23}, 3, 14, h)
			if marvlib.API.InputMousePressed() {
				t.selectedHoriz = h
			}
		} else {
			areaText.StringToMap(image.Point{X: 13 + 3 + (i * 3), Y: 23}, 7, 10, h)
		}
	}
}

func (t *Text) Update() {
	marvlib.API.SpritesGet(SpriteMousePointer).ChangePos(image.Rectangle{
		Min: marvlib.API.InputMousePos(),
		Max: image.Point{X: 32, Y: 32},
	})

	if t.plotInput {
		t.updatePlotUI()
	}
}

const (
	LIKE_HEADSTONE_CATS    = "They left their money to cats."
	LIKE_HEADSTONE_DOGS    = "They left their money to dogs."
	LIKE_HEADSTONE_CROSSES = "They really liked big crosses."
	LIKE_HEADSTONE_WOOD    = "They really liked wood."
	LIKE_HEADSTONE_WORDS   = "They really liked words."

	LIKE_BODY_TALL  = "They were tall."
	LIKE_BODY_SHORT = "They were short."

	LIKE_WORE_GLASSES = "They wore stylish glasses."
	LIKE_WORE_HAT     = "They famously wore a hat."
	LIKE_WORE_BEARD   = "They were bearded."

	LIKE_MARKER_FLOWERS   = "They loved flowers."
	LIKE_MARKER_NOFLOWERS = "They hated flowers."
)

var (
	Likes = [][]string{
		{LIKE_HEADSTONE_CATS, LIKE_HEADSTONE_DOGS, LIKE_HEADSTONE_CROSSES, LIKE_HEADSTONE_WOOD, LIKE_HEADSTONE_WORDS},
		{LIKE_BODY_TALL, LIKE_BODY_SHORT},
		{LIKE_WORE_GLASSES, LIKE_WORE_HAT, LIKE_WORE_BEARD},
		{LIKE_MARKER_NOFLOWERS, LIKE_MARKER_FLOWERS},
	}
	Relations = []string{
		"ancient ancestor",
		"great aunt",
		"great uncle",
		"great grandfather",
		"second cousin",
		"favourite barber",
		"favourite celebrity",
		"old teacher",
		"old neighbour",
		"evil twin",
		"old accountant",
		"old dentist",
		"disgraced plastic surgeon",
	}
)

type Hitbox struct {
	image.Rectangle
}

func (hb Hitbox) IsHit(screenpoint image.Point) bool {
	return screenpoint.Add(camera.GetAsPoint()).In(hb.Rectangle)
}

func (hb Hitbox) IsHitNoCameraOffset(screenpoint image.Point) bool {
	return screenpoint.In(hb.Rectangle)
}

type Grave struct {
	MapX         int
	MapY         int
	GridX        string
	GridY        string
	Relation     string
	Likes        []string
	Hitbox       Hitbox
	clickcounter int
	DigDepth     int

	headStone int
	body      int
	wore      [2]int
	marker    int
}

func NewGrave(x, y int, gridX, gridY string) *Grave {
	g := &Grave{
		DigDepth: 10,
		MapX:     x,
		MapY:     y,
		GridX:    gridX,
		GridY:    gridY,
		Hitbox: Hitbox{
			Rectangle: image.Rectangle{
				Min: image.Point{X: x * 8, Y: y * 8},
				Max: image.Point{X: (x * 8) + (4 * 8), Y: (y * 8) + (10 * 8)},
			},
		},
	}
	g.generate()
	return g
}

var (
	DiggingPhrases = [][]string{
		{""},
		{"<digs> Let's dig this!", "<digs> Open says me!", "<digs> I'm a tomb spader!", "<digs> I'm a whom raider!", "<digs> Plot twist!"},
		{"<digs more>"},
		{"<digs more>", "<digs more> Phew!"},
		{"<digs yet more>"},
		{"<digs more>", "<digs more> Hit a stone!"},
		{"<digs even more>"},
		{"<digs more> Oof!", "<digs more> Feels like clay!"},
		{"<digs more> Almost there!", "<digs more> Regulation depth!"},
		{"*thunk* *thunk*", "*thud* *thud*", "*tink* *tink*"},
	}

	MapLikeToHeadStone = map[string]int{
		LIKE_HEADSTONE_CATS:    5 * 4,
		LIKE_HEADSTONE_DOGS:    6 * 4,
		LIKE_HEADSTONE_CROSSES: 3 * 4,
		LIKE_HEADSTONE_WOOD:    3 * 4,
		LIKE_HEADSTONE_WORDS:   7 * 4,
	}

	MapLikeToBody = map[string][]int{
		LIKE_BODY_TALL:  {2 * 4, 4 * 4},
		LIKE_BODY_SHORT: {3 * 4, 5 * 4},
	}

	MapLikeToWore = map[string][2]int{
		LIKE_WORE_BEARD:   [2]int{6 * 4, 3 * 4}, // DIFF ROW!!!
		LIKE_WORE_GLASSES: [2]int{6 * 4, 2 * 4},
		LIKE_WORE_HAT:     [2]int{7 * 4, 2 * 4},
	}

	MapLikeToMarker = map[string][]int{
		LIKE_MARKER_FLOWERS:   {6 * 4, 7 * 4},
		LIKE_MARKER_NOFLOWERS: {33},
	}
)

func (g *Grave) Dig() {
	g.clickcounter++
	text.VisitorSay("", "")
	if g.clickcounter < g.DigDepth {
		text.PlayerSay(DiggingPhrases[g.clickcounter][rand.Intn(len(DiggingPhrases[g.clickcounter]))])
		camera.Shake()
	} else if g.clickcounter == g.DigDepth {
		g.drawOpenGrave()
		text.PlayerSay("That's got you, " + g.GridY + g.GridX)
	} else {
		text.PlayerSay("Lookin' good there, " + g.GridY + g.GridX)
	}
}

func (g *Grave) getHeadstoneFromLikes() {
	g.headStone = 4 * 4 // TODO: random if no like???
	for _, like := range g.Likes {
		if idx, found := MapLikeToHeadStone[like]; found {
			g.headStone = idx
		}
	}
}
func (g *Grave) getBodyFromLikes() {
	randBody := MapLikeToBody[LIKE_BODY_TALL]   // default to tall
	g.body = randBody[rand.Intn(len(randBody))] // default to random as it doesn't matter

	for _, like := range g.Likes {
		if options, found := MapLikeToBody[like]; found {
			g.body = options[rand.Intn(len(options))]
		}
	}
}
func (g *Grave) getWoreFromLikes() {
	g.wore = [2]int{32, 32} // TODO: random if no like???
	for _, like := range g.Likes {
		if coords, found := MapLikeToWore[like]; found {
			g.wore = coords
		}
	}
}
func (g *Grave) getMarkerFromLikes() {
	g.marker = 32 // TODO: random if no like???
	for _, like := range g.Likes {
		if options, found := MapLikeToMarker[like]; found {
			g.marker = options[rand.Intn(len(options))]
		}
	}
}

func (g *Grave) generate() {
	g.generateRelation()
	g.generateLikes()
	g.getHeadstoneFromLikes()
	g.getBodyFromLikes()
	g.getWoreFromLikes()
	g.getMarkerFromLikes()
}

func (g *Grave) generateRelation() {
	g.Relation = Relations[rand.Intn(len(Relations))]
}

func (g *Grave) generateLikes() {
	// shuffle and reduce to 3 rows of exclusive options
	l := Likes
	rand.Shuffle(len(l), func(i, j int) {
		l[i], l[j] = l[j], l[i]
	})
	l = l[:3]
	// pick one exclusive option from each row
	g.Likes = []string{}
	for _, options := range l {
		g.Likes = append(g.Likes, options[rand.Intn(len(options))])
	}
}

func (g *Grave) drawCommon(dstX int, dstY int, srcX int, srcY int, w int, h int) {
	offset := image.Point{X: dstX, Y: dstY}
	pos := image.Point{}
	for pos.Y = 0; pos.Y < h; pos.Y++ {
		for pos.X = 0; pos.X < w; pos.X++ {
			area.Set(pos.Add(offset), uint8(srcX+pos.X), uint8(srcY+pos.Y), 0, 0)
		}
	}
}

func (g *Grave) drawCommonOverlay(dstX int, dstY int, srcX int, srcY int, w int, h int) {
	offset := image.Point{X: dstX, Y: dstY}
	pos := image.Point{}
	for pos.Y = 0; pos.Y < h; pos.Y++ {
		for pos.X = 0; pos.X < w; pos.X++ {
			areaOverlay.Set(pos.Add(offset), uint8(srcX+pos.X), uint8(srcY+pos.Y), 0, 0)
		}
	}
}

func (g *Grave) drawHeadstone() {
	g.drawCommon(g.MapX, g.MapY, g.headStone, 0, 4, 4) // headstone
}

func (g *Grave) drawMarker() {
	g.drawCommonOverlay(g.MapX, g.MapY+3, g.marker, 4, 4, 4)
}

func (g *Grave) drawWore() {
	g.drawCommonOverlay(g.MapX, g.MapY+2, g.wore[0], g.wore[1], 4, 6)
}

func (g *Grave) drawClosedGrave() {
	g.drawHeadstone()
	r := 8 + (rand.Intn(4) * 4)                // random closed grave
	g.drawCommon(g.MapX, g.MapY+4, r, 4, 4, 6) // grave
	g.drawMarker()
}

func (g *Grave) drawOpenGrave() {
	g.drawHeadstone()
	g.drawCommon(g.MapX, g.MapY+4, g.body, 10, 4, 6) // grave
	g.drawWore()
}

var (
	NPCId = SpriteVisitor1
)

type Visitor struct {
	Pos      image.Point
	SpriteId int
	Grave    *Grave
	Hitbox   Hitbox

	w int
	h int

	done bool
}

func NewVisitor(grave *Grave) *Visitor {
	v := &Visitor{
		SpriteId: NPCId,
		Grave:    grave,
		w:        4 * 8,
		h:        8 * 8,
	}
	NPCId++
	marvlib.API.SpritesGet(v.SpriteId).Show(GfxBankPeople, areaPeople)
	marvlib.API.SpritesGet(v.SpriteId).ChangeViewport(image.Point{X: (3 + rand.Intn(4)) * 32, Y: 0})

	v.Pos.X = 128 + 32 + ((v.SpriteId - SpriteVisitor1) * (40))
	v.Pos.Y = 54

	marvlib.API.SpritesGet(v.SpriteId).ChangePos(image.Rectangle{
		Min: v.Pos,
		Max: image.Point{X: v.w, Y: v.h},
	})
	marvlib.API.SpritesGet(v.SpriteId).SetSortIdx(v.Pos.Y)
	marvlib.API.SpritesSort()
	v.updateHitbox()
	return v
}

func (v *Visitor) updateHitbox() {
	v.Hitbox = Hitbox{
		Rectangle: image.Rectangle{
			Min: v.Pos,
			Max: v.Pos.Add(image.Point{X: v.w, Y: v.h}),
		},
	}
}

func (v *Visitor) ChoosePlot(gridV, gridH string) {
	selectedPlot := gridV + gridH
	actualPlot := v.Grave.GridY + v.Grave.GridX

	// fmt.Println(selectedPlot, actualPlot)

	if selectedPlot == actualPlot {
		text.VisitorSay(
			"It's so nice to see them\nagain! Although with perhaps\na touch more clarity than\nexpected...\n\nThank you!",
			fmt.Sprintf("Happy Visitor %d", (v.SpriteId-SpriteVisitor1)+1),
		)
		camera.TargetX = float64(player.X)
		camera.TargetY = float64(player.Y)

		v.Pos.X = (v.Grave.MapX + 3) * 8
		v.Pos.Y = (v.Grave.MapY + 2) * 8
		v.Hitbox = Hitbox{} // make untouchable

	} else {
		text.VisitorSay(
			"You couldn't be more wrong!\nI'm off in a huff!\nTwo, if I can manage it!",
			fmt.Sprintf("Angry Visitor %d", (v.SpriteId-SpriteVisitor1)+1),
		)
		camera.TargetX = float64(player.X)
		camera.TargetY = float64(player.Y)
		v.Pos.Y = -128
		v.Hitbox = Hitbox{} // make untouchable
	}

	marvlib.API.SpritesGet(v.SpriteId).SetSortIdx(v.Pos.Y)
	marvlib.API.SpritesSort()

	v.done = true
}

func (v *Visitor) Update() {
	marvlib.API.SpritesGet(v.SpriteId).ChangePos(image.Rectangle{
		Min: v.Pos.Sub(camera.GetAsPoint()),
		Max: image.Point{X: v.w, Y: v.h},
	})
}

type Visitors struct {
	Visitors       []*Visitor
	currentVisitor *Visitor
}

func NewVisitors(numVisitors int) *Visitors {
	NPCId = SpriteVisitor1

	v := &Visitors{}

	for i := 0; i < numVisitors; i++ {
		v.AddVisitor()
	}

	return v
}

func (v *Visitors) AddVisitor() {
	// randomly pick a grave (that hasn't been chosen yet??? or could be same person so no matter!!!)
	randomGrave := graveyard.graves[rand.Intn(len(graveyard.graves))]
	newVisitor := NewVisitor(randomGrave)
	v.Visitors = append(v.Visitors, newVisitor)
}

func (v *Visitors) GetClickedVisitor(screenpoint image.Point) *Visitor {
	for _, visitor := range v.Visitors {
		if visitor.Hitbox.IsHit(screenpoint) {
			return visitor
		}
	}
	return nil
}

func (v *Visitors) Update() {
	doneCount := 0
	for _, visitor := range v.Visitors {
		visitor.Update()
		if visitor.done {
			doneCount++
			if doneCount == len(v.Visitors) {
				nextLevel()
			}
		}
	}
}

func nextLevel() {
	for _, visitor := range visitors.Visitors {
		visitor.done = false // prevent success looping!
	}
	time.AfterFunc(6*time.Second, func() {
		lvlNum++
		graveyard = &Graveyard{}
		graveyard.Setup(lvlNum)
		visitors = NewVisitors(5)
		player.flagDoneSomeDigging = false
		player.flagTalkedToVisitors = false
		text.VisitorSay("<The next day...>", fmt.Sprintf("Level %d", lvlNum))
		text.PlayerSay("")
		text.PlayerSay("Another new day dawns!\nHas the graveyard got bigger!?\nMust be my eyes...")
	})
}

var (
	camera    *Camera
	player    *Player
	graveyard *Graveyard
	text      *Text
	visitors  *Visitors
	lvlNum    = 1

	areaUnderlay     marvtypes.MapBankArea
	area             marvtypes.MapBankArea
	areaOverlay      marvtypes.MapBankArea
	areaMousePointer marvtypes.MapBankArea
	areaPeople       marvtypes.MapBankArea
	areaTitles       marvtypes.MapBankArea
	areaText         marvtypes.MapBankArea
)

func setupAreas() {
	areaPeople = marvlib.API.MapBanksGet(MapBankPeople).AllocArea(image.Point{X: 32, Y: 32})
	areaTitles = marvlib.API.MapBanksGet(MapBankGraveyard).AllocArea(image.Point{X: 32, Y: 32})
	areaOverlay = marvlib.API.MapBanksGet(MapBankGraveyard).AllocArea(image.Point{X: 64, Y: 249})
	areaUnderlay = marvlib.API.MapBanksGet(MapBankGraveyard).AllocArea(image.Point{X: 64, Y: 249})
	area = marvlib.API.MapBanksGet(MapBankGraveyard).AllocArea(image.Point{X: 64, Y: 249})
	areaMousePointer = marvlib.API.MapBanksGet(MapBankGraveyard).AllocArea(image.Point{X: 4, Y: 4})
	areaText = marvlib.API.MapBanksGet(MapBankText).AllocArea(image.Point{X: 54, Y: 25}) // full screen in default 6x8 font

	// people
	pos := image.Point{}
	for pos.Y = 0; pos.Y < 32; pos.Y++ {
		for pos.X = 0; pos.X < 32; pos.X++ {
			areaPeople.Set(pos, uint8(pos.X), uint8(pos.Y), 5, 5)
			areaTitles.Set(pos, uint8(pos.X), uint8(pos.Y), 5, 5)
		}
	}
}

func setupGame() {
	camera = &Camera{X: 0, Y: 0}
	player = NewPlayer()
	graveyard = &Graveyard{} // we need global refs to this before it sets up, hence the two set
	graveyard.Setup(lvlNum)
	text = NewText()
	visitors = NewVisitors(5)

	text.PlayerSay("I must talk to the visitors.\nAfter all, I'm here to help\nfind who they're looking for.")

	mode = MODE_GAME
}

func updateGame() {
	camera.Update()
	graveyard.Update()
	player.Update()
	text.Update()
	visitors.Update()
}

func setupTitles() {
	pos := image.Point{}
	for pos.Y = 0; pos.Y < 32; pos.Y++ {
		for pos.X = 0; pos.X < 32; pos.X++ {
			areaTitles.Set(pos, uint8(pos.X), uint8(pos.Y), 0, 0)
		}
	}

	marvlib.API.SpritesGet(SpriteGraveyard).ChangePos(titleScreenRect)
	marvlib.API.SpritesGet(SpriteGraveyard).Show(GfxBankTitles, areaTitles)

	mode = MODE_TITLE_SCREEN
}

func updateTitles() {
	if marvlib.API.InputMousePressed() {
		mode = MODE_GAME_SETUP
	}
}

func Start() {
	setupAreas()
	// marv.ModBanks[0].Play()
}

func Update() {
	switch mode {
	case MODE_TITLE_SCREEN_SETUP:
		setupTitles()
		// marv.MapBanks[MapBankGraveyard].Save("foo.map")
		// marv.MapBanks[MapBankGraveyard].Load("foo.map")
		// marv.Sprites[SpriteGraveyard].Show(areaTitles)
	case MODE_TITLE_SCREEN:
		updateTitles()
	case MODE_GAME_SETUP:
		setupGame()
	case MODE_GAME:
		updateGame()
	}
}
