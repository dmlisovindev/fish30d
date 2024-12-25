package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2/colorm"

	"github.com/hajimehoshi/ebiten/v2/inpututil"

	_ "image/png"
	"log"

	// "github.com/fish30d/fish30d/fishes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	//go:embed resources/playerfish.png
	playerImage []byte
	//go:embed resources/bass.png
	bassImage []byte
	//go:embed resources/goldfish.png
	goldfishImage []byte
	//go:embed resources/puffer.png
	pufferImage []byte
	//go:embed resources/shark.png
	sharkImage []byte
	//go:embed resources/jellyfish.png
	jellyImage []byte
	//go:embed resources/TheGoodMonolith.ttf
	fixedsys []byte
	//go:embed resources/AquaWow.otf
	aquawow []byte
	quotes  = []string{
		"YOU DIED",
		"You tried.",
		"You're fry-ed.",
		"Try again.",
		"You've got a valuable lesson.",
		"Press Spacebar to restart.",
		"Be careful next time.",
		"So it goes.",
		"It happens.",
		"Too greedy.",
		"You just went belly up.",
		"Well, at least give them indigestion.",
		"You still have two ways out.",
		"Some fish can swallow prey 3 times their own size. Alas, you're not one of those.",
		"There's always a bigger fish.",
		"Don't bite more than you can swallow.",
		"That's it, no more Bill, his greed got him killed.",
		"Oops, someone got greedy.",
		"Eat the fish smaller than you, or be eaten by a bigger fish.",
		"Use the keyboard or the mouse to move, whatever suits you best.",
		"If you're cornered, press Spacebar or RMB to dodge to back or front plane.",
		"Press H or click the little arrow in the main menu to hide it and just relax.",
		"Press P or Enter to pause the game.",
		"Sharks attack anything they see as a good meal, for better or worse.",
		"Goldfish will try to escape with a dodge and a dash.",
		"Pufferfish are easily scared and puff up to make themselves inedible... or more delicious.",
		"The bass will run away when threatened, but will always try to sneak back.",
		"The jellyfish is brainless but not harmless. Still, just as edible as the other fish.",
		"Remember, objects further to the back are bigger than they appear.",
		"If you want more - or less - challenge, go to Options and play around.",
		"The bigger the fish, the better the score, but don't get too greedy.",
		"If the fish isn't afraid of you, that may be for a reason.",
		"GULP",
	}
)

const (
	title           = "FISH 3.0D"
	screenWidth     = 1920
	screenHeight    = 1080
	fishCount       = 10
	maxPlanes       = 2
	gameRunning     = 0
	gameOver        = 2
	gameVictory     = 3
	gameMenu        = 4
	gameOptionsMenu = 5
)

func getFishType(index, fishesCount float64) string {
	switch {
	case index == 0:
		return "jelly"
	case index < 0.35*fishesCount:
		return "bass"
	case index < 0.65*fishesCount:
		return "goldfish"
	case index < 0.85*fishesCount:
		return "puffer"
	default:
		return "shark"
	}

}

type Fish struct {
	Size                float64
	HalfWidth           float64
	HalfHeight          float64
	Scale               float64
	X                   float64
	Y                   float64
	SpeedX              float64
	SpeedY              float64
	FacingLeft          bool
	Plane               float64
	Dead                bool
	Type                string
	FrictionCoefficient float64
	GraphUpdated        bool
	Cooldown            float64
	Image               *ebiten.Image
	DrawOptions         *colorm.DrawImageOptions
	Colorm              *colorm.ColorM
	game                *Game
}

type PlayerFish struct {
	Fish
}

func (fish *Fish) CooldownTick() {
	if fish.Cooldown == 0 {
		return
	}
	fish.Cooldown--
	switch fish.Type {
	case "puffer":
		speedFactor := 6.25
		switch {
		case fish.Cooldown < 0:
			fish.Cooldown = 60 * 20
			fish.SpeedX /= speedFactor
			fish.SpeedY /= speedFactor
		case fish.Cooldown > 60*18.5:
			fish.SetSize(fish.Size * 1.01)
		case fish.Cooldown > 60*15 && fish.Cooldown <= 60*16.5:
			fish.SetSize(fish.Size / 1.01)
		case fish.Cooldown == 60*15:
			fish.SpeedX *= speedFactor
			fish.SpeedY *= speedFactor
		}
	case "goldfish":
		speedFactor := 3.0
		switch {
		case fish.Cooldown < 0:
			fish.Cooldown = 5 * 60
			fish.SpeedX /= 10
			fish.SpeedY /= 10
		case fish.Cooldown == 4.5*60:
			fish.SwitchPlane()
			fish.SpeedX *= speedFactor * 10
			fish.SpeedY *= speedFactor * 10
		case fish.Cooldown == 4*60:
			fish.SpeedX /= speedFactor
			fish.SpeedY /= speedFactor

		}
	case "shark":
		switch {
		case fish.Cooldown < 0:
			fish.Cooldown = 10 * 60
		case fish.Cooldown == 9*60:
			fish.SpeedX /= -4
			fish.SpeedY /= -4
			fish.FacingLeft = fish.SpeedX < 0
			fish.GraphUpdated = true

		}
	case "bass":
		turn := math.Mod(fish.Size, 2)
		switch {
		case fish.Cooldown < 0:
			fish.Cooldown = 10 * 60
		case fish.Cooldown == 9*60:
			if turn == 0 {
				fish.SpeedX = -1 * fish.SpeedX
			} else {
				fish.SpeedY = -1 * fish.SpeedY
			}
			fish.FacingLeft = fish.SpeedX < 0
			fish.GraphUpdated = true
		case fish.Cooldown == 8*60:
			if turn == 0 {
				fish.SpeedY = -1 * fish.SpeedY
			} else {
				fish.SpeedX = -1 * fish.SpeedX
			}
			fish.FacingLeft = fish.SpeedX < 0
			fish.GraphUpdated = true
		case fish.Cooldown == 7*60:
			fish.SpeedX /= 2
			fish.SpeedY /= 2
		}
	}

}

func (fish *Fish) Die() bool {
	if !fish.Dead {
		fish.Dead = true
		fish.SpeedX = 0
		fish.SpeedY = 0
		fish.Cooldown = 0
		fish.GraphUpdated = true
		return true
	}
	return false
}

func (fish *Fish) Draw() {
	if fish.GraphUpdated {
		fish.GraphReset()
	} else {
		fish.DrawOptions.GeoM.Translate(fish.SpeedX, fish.SpeedY)
	}
	colorm.DrawImage(fish.game.screen, fish.Image, *fish.Colorm, fish.DrawOptions)
}

func (fish *Fish) GraphReset() {
	fish.DrawOptions.Blend = ebiten.BlendSourceOver
	fish.DrawOptions.GeoM.Reset()
	fish.Colorm.Reset()
	flipX, flipY := float64(1), float64(1)
	if fish.FacingLeft {
		flipX = float64(-1)
	}
	if fish.Dead {
		flipY = float64(-1)
		fish.Colorm.ChangeHSV(0, 0, 2)
		fish.Colorm.Scale(1, 1, 1, 0.25)
	} else {
		if fish.Plane > 0 {
			fish.DrawOptions.Blend = ebiten.BlendXor
		}
	}
	fish.DrawOptions.Filter = ebiten.FilterLinear
	fish.DrawOptions.GeoM.Scale(fish.Scale*flipX, fish.Scale*flipY)
	fish.DrawOptions.GeoM.Translate(fish.X-flipX*fish.HalfWidth, fish.Y-flipY*fish.HalfHeight)
}

func (fish *PlayerFish) Hit(target *Fish) {
	if !target.Dead && !fish.Dead && fish.Overlap(target) {
		if fish.Size < target.Size {
			fish.Die()
			fish.game.VibrateGamepadHeavy()
		} else {
			if target.Die() {
				fish.SetSize(fish.Size + 1)
				fish.game.UpdateScore(target.Size)
				fish.game.VibrateGamepadQuick()
				if fish.HalfWidth*2 > fish.game.screenWidth {
					fish.game.Win()
				}
			}

		}
	}
}

func (fish *PlayerFish) Hunt(targets []Fish) {
	for i, _ := range targets {
		if fish.game.fishReactionsEnabled {
			targets[i].ProximityAlert(fish)
		}
		fish.Hit(&targets[i])
	}
}

func (fish *Fish) Init(g *Game, fishtype string) {
	fish.game = g
	fish.Type = fishtype
	fish.InitImage()

}

func (fish *PlayerFish) Init(g *Game) {
	fish.game = g
	fish.Type = "player"
	fish.InitImage()
}

func (fish *Fish) InitImage() {
	if fish.Image != nil {
		fish.Image.Dispose()
	}
	fish.Image = ebiten.NewImageFromImage(fish.game.preloadedImages[fish.Type])
	fish.DrawOptions = new(colorm.DrawImageOptions)
	fish.Colorm = new(colorm.ColorM)
}

func (fish *Fish) IsOutOfBounds() (isOut, vertical bool) {
	if fish.Cooldown != 0 {
		return false, false
	}
	horizontal := !(fish.X >= -fish.HalfWidth && fish.X <= fish.game.screenWidth+fish.HalfWidth)
	vertical = !(fish.Y >= -fish.HalfHeight && fish.Y <= fish.game.screenHeight+fish.HalfHeight)
	isOut = horizontal || vertical
	return
}

func (fish *PlayerFish) IsOutOfBounds() (isOut, vertical bool) {
	horizontal := (fish.SpeedX < 0 && fish.X < fish.HalfWidth) || (fish.SpeedX > 0 && fish.X > fish.game.screenWidth-fish.HalfWidth)
	vertical = (fish.SpeedY < 0 && fish.Y < fish.HalfHeight) || (fish.SpeedY > 0 && fish.Y > fish.game.screenHeight+fish.HalfHeight)
	isOut = horizontal || vertical
	return
}

func (fish *Fish) Move() {
	fish.Swim(0, 0) //look an owl
	if out, _ := fish.IsOutOfBounds(); out {
		fish.Randomize()
	}
	fish.CooldownTick()

}

func (fish *PlayerFish) Move() {
	driveX, driveY := fish.ReadInput()
	fish.Swim(driveX, driveY)
	if out, vertical := fish.IsOutOfBounds(); out {
		fish.Rebound(vertical)
	}
	fish.Hunt(fish.game.fish)
}

func (fish *Fish) Overlap(target *Fish) bool {
	if fish.Plane != target.Plane {
		return false
	}
	fRectangle := image.Rect(int(fish.X-fish.HalfWidth), int(fish.Y-fish.HalfHeight), int(fish.X+fish.HalfWidth), int(fish.Y+fish.HalfHeight))
	tRectangle := image.Rect(int(target.X-target.HalfWidth), int(target.Y-target.HalfHeight), int(target.X+target.HalfWidth), int(target.Y+target.HalfHeight))
	intersection := fRectangle.Intersect(tRectangle)
	if !intersection.Empty() {
		matrix, tmatrix := fish.DrawOptions.GeoM, target.DrawOptions.GeoM
		matrix.Invert()
		tmatrix.Invert()
		for y := intersection.Min.Y; y < intersection.Max.Y; y++ {
			for x := intersection.Min.X; x < intersection.Max.X; x++ {
				x0, y0 := matrix.Apply(float64(x), float64(y))
				tx0, ty0 := tmatrix.Apply(float64(x), float64(y))
				_, _, _, a := fish.Image.At(int(x0), int(y0)).RGBA()
				_, _, _, ta := target.Image.At(int(tx0), int(ty0)).RGBA()
				if a != 0 && ta != 0 {
					return true
				}
			}
		}
	}
	return false
}

func (fish *Fish) ProximityAlert(attacker *PlayerFish) {
	distance := math.Sqrt(math.Pow(fish.X-attacker.X, 2) + math.Pow(fish.Y-attacker.Y, 2))
	if fish.Dead || fish.Cooldown != 0 || fish.Plane != attacker.Plane || distance > 4*(fish.HalfWidth+attacker.HalfWidth) {
		return
	}

	if (fish.Type == "puffer" && fish.Size < attacker.Size*2) || (fish.Type == "goldfish" && fish.Size <= attacker.Size) {
		fish.Cooldown = -1
	}

	if fish.Type == "bass" && fish.Size <= attacker.Size {
		fish.SpeedX = math.Copysign(2*fish.SpeedX, fish.X-attacker.X)
		fish.SpeedY = math.Copysign(2*fish.SpeedY, fish.Y-attacker.Y)
		fish.FacingLeft = fish.SpeedX < 0
		fish.GraphUpdated = true
		fish.Cooldown = -1

	}

	if fish.Type == "shark" && attacker.Size > fish.Size/2 && attacker.Size < fish.Size*1.5 {
		fish.Cooldown = -1
		fish.SpeedX, fish.SpeedY = (attacker.X-fish.X)/90, (attacker.Y-fish.Y)/90
		fish.FacingLeft = fish.SpeedX < 0
		fish.GraphUpdated = true
	}
}

func (fish *Fish) Randomize() {
	fish.Dead = false
	fish.Cooldown = 0
	fish.Plane = float64(rand.Intn(int(fish.game.planeCount)))
	fish.SetSize(float64(rand.Intn(int(fish.game.fishSizeCap)-5) + 5))
	fish.SpeedX = float64(rand.Intn(3) + 1)
	fish.SpeedY = float64(rand.Intn(3) - 1)
	reverse := rand.Intn(2)
	if reverse == 1 {
		fish.SpeedX *= -1
	}
	if fish.Type == "jelly" {
		fish.SpeedY = fish.SpeedX
		fish.SpeedX = 0
		fish.X = float64(rand.Intn(int(fish.game.screenWidth)))
		if fish.SpeedY < 0 {
			fish.Y = fish.game.screenHeight + fish.HalfHeight - 1
		} else {
			fish.Y = 1 - fish.HalfHeight
		}
	} else {
		fish.Y = float64(rand.Intn(int(fish.game.screenHeight)))
		if fish.SpeedX < 0 {
			fish.X = fish.game.screenWidth + fish.HalfWidth - 1
		} else {
			fish.X = 1 - fish.HalfWidth
		}
	}
	fish.SpeedX, fish.SpeedY = fish.SpeedX*fish.game.fishSpeedModifier, fish.SpeedY*fish.game.fishSpeedModifier
	fish.FacingLeft = fish.SpeedX < 0
	fish.GraphUpdated = true

}

func (fish *PlayerFish) ReadInput() (driveX, driveY float64) {
	if fish.Dead {
		return
	}
	driveX, driveY = 0, 0

	if isAnyOfKeysPressed(false, ebiten.KeyW, ebiten.KeyArrowUp) {
		driveY -= 1
	}
	if isAnyOfKeysPressed(false, ebiten.KeyS, ebiten.KeyArrowDown) {
		driveY += 1
	}
	if isAnyOfKeysPressed(false, ebiten.KeyA, ebiten.KeyArrowLeft) {
		driveX -= 1
	}
	if isAnyOfKeysPressed(false, ebiten.KeyD, ebiten.KeyArrowRight) {
		driveX += 1
	}
	if isAnyOfKeysPressed(true, ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2) || fish.game.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonRightBottom) {
		fish.SwitchPlane()
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		jx, jy := ebiten.CursorPosition()
		mx, my := float64(jx)-fish.X, float64(jy)-fish.Y
		if math.Hypot(mx, my) >= fish.HalfHeight {
			driveX, driveY = mx, my
		}
	}
	if fish.game.debugEnabled {

		if isAnyOfKeysPressed(false, ebiten.KeyPageUp) {
			fish.SetSize(fish.Size + 1)
		}
		if isAnyOfKeysPressed(false, ebiten.KeyPageDown) {
			fish.SetSize(fish.Size - 1)
		}
		if isAnyOfKeysPressed(false, ebiten.KeyDelete) {
			fish.Die()
		}
	}

	if ebiten.IsStandardGamepadLayoutAvailable(fish.game.gamepadId) {
		gx, gy := ebiten.StandardGamepadAxisValue(fish.game.gamepadId, ebiten.StandardGamepadAxisLeftStickHorizontal), ebiten.StandardGamepadAxisValue(fish.game.gamepadId, ebiten.StandardGamepadAxisLeftStickVertical)
		if math.Hypot(gx, gy) > 0.25 {
			driveX, driveY = gx, gy
		}
	}
	if driveAbs := math.Hypot(driveX, driveY); driveAbs > 1 {
		driveX, driveY = driveX/driveAbs, driveY/driveAbs
	}
	if driveX != 0 || driveY != 0 {
		fish.FacingLeft = driveX < 0
		fish.GraphUpdated = true
	}
	return
}

func (fish *PlayerFish) Rebound(vertical bool) {
	switch vertical {
	case false:
		fish.X -= fish.SpeedX
		fish.SpeedX *= -0.5
	case true:
		if fish.Dead {
			fish.game.GameOver()
			return
		}
		fish.Y -= fish.SpeedY
		fish.SpeedY *= -0.5
	}
}

func (fish *PlayerFish) Reset() {
	fish.Dead = false
	fish.Plane = 0
	fish.SetSize(10)
	fish.X = float64(fish.game.screenWidth) / 2
	fish.Y = float64(fish.game.screenHeight) / 2
	fish.SpeedX, fish.SpeedY = 0, 0
	fish.FrictionCoefficient = 1
	fish.GraphUpdated = true
}

func (fish *Fish) ResizeSprite() {
	fish.Scale = math.Pow(0.75, fish.Plane) * fish.Size / 64
	actualSize := fish.Image.Bounds()
	fish.HalfWidth, fish.HalfHeight = fish.Scale*float64(actualSize.Dx())/2, fish.Scale*float64(actualSize.Dy())/2
	fish.GraphUpdated = true
}

func (fish *Fish) SetSize(newSize float64) {
	oldSize := fish.Size
	fish.Size = newSize
	if fish.Size < 1 {
		fish.Size = oldSize
		return
	}
	fish.ResizeSprite()

}

func (fish *Fish) Swim(driveX, driveY float64) {
	var accX, accY float64
	if fish.Dead {
		accX, accY = 0, -0.2
	} else {
		accX = driveX*fish.game.playerAcceleration + fish.FrictionCoefficient*fish.SpeedX*fish.game.playerDeceleration
		accY = driveY*fish.game.playerAcceleration + fish.FrictionCoefficient*fish.SpeedY*fish.game.playerDeceleration
	}
	fish.SpeedX += accX
	fish.SpeedY += accY
	fish.X += fish.SpeedX
	fish.Y += fish.SpeedY
}

func (fish *Fish) SwitchPlane() {
	fish.Plane = math.Mod(fish.Plane+1, fish.game.planeCount)
	fish.ResizeSprite()
}

type MenuItem struct {
	title    string
	x, y, h  float64
	selector int
	fontFace *text.GoTextFace
	values   []float64
	titles   []string
}

func (m *MenuItem) Draw(screen *ebiten.Image, active bool) {

	itemText := m.title
	if len(m.titles) > 0 {
		itemText += ": " + m.titles[m.selector]
	}
	op := &text.DrawOptions{}
	if active {
		op.ColorScale.ScaleWithColor(color.RGBA{255, 128, 0, 255})
	}
	op.GeoM.Translate(m.x, m.y+(m.h/2)-(m.fontFace.Size/2))
	text.Draw(screen, itemText, m.fontFace, op)
}

func (m *MenuItem) DetectHover() bool {
	_, y := ebiten.CursorPosition()
	fy := float64(y)
	return fy >= m.y && fy < m.y+m.h
}

func (m *MenuItem) CycleRight() {
	m.selector = (m.selector + 1) % len(m.values)
}

func (m *MenuItem) ShiftRight() bool {
	s := int(math.Min(float64(m.selector+1), float64(len(m.values)-1)))
	if m.selector == s {
		return false
	}
	m.selector = s
	return true
}

func (m *MenuItem) ShiftLeft() bool {
	s := int(math.Max(float64(m.selector-1), 0))
	if m.selector == s {
		return false
	}
	m.selector = s
	return true
}

func (m *MenuItem) GetValue() float64 {
	return m.values[m.selector]
}

type Game struct {
	activeMenuIndex      int
	background           color.Color
	debugEnabled         bool
	eaten                float64
	fancyFontSource      *text.GoTextFaceSource
	fishPerPlane         float64
	fishSpeedModifier    float64
	fishSizeCap          float64
	fishReactionsEnabled bool
	fishStaticArray      [50]Fish
	fish                 []Fish
	fontSizes            map[string]float64
	gamepadId            ebiten.GamepadID
	gameState            int
	highScore            float64
	mainMenu             []MenuItem
	menuHidden           bool
	mostEaten            float64
	optionsMenu          []MenuItem
	paused               bool
	plainFontSource      *text.GoTextFaceSource
	planeCount           float64
	playerAcceleration   float64
	playerDeceleration   float64
	playerFish           PlayerFish
	preloadedImages      map[string]image.Image
	prevCurX             int
	prevCurY             int
	randomQuote          string
	score                float64
	screen               *ebiten.Image
	screenHeight         float64
	screenWidth          float64
	totalFishCount       int
}

func (g *Game) ApplyOptions() {
	g.planeCount = g.optionsMenu[0].GetValue()
	g.fishPerPlane = g.optionsMenu[1].GetValue()
	g.fishSpeedModifier = g.optionsMenu[2].GetValue()
	g.fishSizeCap = g.optionsMenu[3].GetValue()
	g.fishReactionsEnabled = g.optionsMenu[4].GetValue() == 1
	ebiten.SetFullscreen(g.optionsMenu[5].GetValue() == 1)
	g.GenerateFish()
	for i, _ := range g.fish {
		g.fish[i].Randomize()
	}
}

func (g *Game) CreateMenus() {
	face := g.GetFontFace("big", true)
	faceOpt := g.GetFontFace("biggish", true)
	x := 0.2 * g.screenWidth
	y := 0.35 * g.screenHeight
	h := 0.2 * g.screenHeight
	mainMenuItems := []string{
		"PLAY", "Options", "Quit",
	}
	for _, title := range mainMenuItems {
		g.mainMenu = append(g.mainMenu, MenuItem{
			title:    title,
			x:        x,
			y:        y,
			h:        h,
			fontFace: face,
		})
		y += h
	}
	x = 0.2 * g.screenWidth
	y = 0.075 * g.screenHeight
	h = 0.12 * g.screenHeight
	g.optionsMenu = append(g.optionsMenu, MenuItem{
		title:    "Game planes",
		x:        x,
		y:        y,
		h:        h,
		fontFace: faceOpt,
		selector: 1,
		titles:   []string{"1", "2"},
		values:   []float64{1, 2},
	})
	y += h
	g.optionsMenu = append(g.optionsMenu, MenuItem{
		title:    "Fish amount",
		x:        x,
		y:        y,
		h:        h,
		fontFace: faceOpt,
		selector: 2,
		titles:   []string{"scarce", "less", "normal", "more", "swarm"},
		values:   []float64{5, 10, 15, 20, 25},
	})
	y += h
	g.optionsMenu = append(g.optionsMenu, MenuItem{
		title: "Fish speed",

		x:        x,
		y:        y,
		h:        h,
		fontFace: faceOpt,
		selector: 1,
		titles:   []string{"slow", "normal", "fast", "frenzy"},
		values:   []float64{0.5, 1, 1.5, 2},
	})
	y += h
	g.optionsMenu = append(g.optionsMenu, MenuItem{
		title:    "Fish max size",
		x:        x,
		y:        y,
		h:        h,
		fontFace: faceOpt,
		selector: 0,
		titles:   []string{"big", "bigger", "biggest"},
		values:   []float64{45, 60, 75},
	})
	y += h
	g.optionsMenu = append(g.optionsMenu, MenuItem{
		title:    "Fish reactions",
		x:        x,
		y:        y,
		h:        h,
		fontFace: faceOpt,
		selector: 1,
		titles:   []string{"off", "on"},
		values:   []float64{0, 1},
	})
	y += h
	g.optionsMenu = append(g.optionsMenu, MenuItem{
		title:    "Fullscreen",
		x:        x,
		y:        y,
		h:        h,
		fontFace: faceOpt,
		selector: 1,
		titles:   []string{"no", "yes"},
		values:   []float64{0, 1},
	})
	y += h
	g.optionsMenu = append(g.optionsMenu, MenuItem{
		title:    "Back",
		x:        x,
		y:        y,
		h:        h,
		fontFace: faceOpt,
	})
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.screen = screen
	screen.Fill(g.background)

	switch g.gameState {
	case gameRunning:
		g.DrawGame()
	case gameOver:
		g.DrawGameOver()
	case gameMenu:
		g.DrawMenu()
	case gameOptionsMenu:
		g.DrawOptions()
	case gameVictory:
		g.DrawVictory()
	}
}

func (g *Game) DrawAllNpcFish() {
	g.DrawPlanesRecursive(g.fish, g.totalFishCount, int(g.planeCount-1))
}

func (g *Game) DrawGame() {
	if !g.playerFish.Dead {
		g.DrawAllNpcFish()
	}
	g.playerFish.Draw()
	if g.debugEnabled {
		ebitenutil.DebugPrint(g.screen, fmt.Sprintf("Fish position (X Y): %0.2f %0.2f Fish Speed (X Y): %0.5f %0.5f Size: %0.0f axis: %0.2f",
			g.playerFish.X, g.playerFish.Y, g.playerFish.SpeedX, g.playerFish.SpeedY, g.playerFish.Size, ebiten.StandardGamepadAxisValue(g.gamepadId, ebiten.StandardGamepadAxisLeftStickHorizontal)))
	}
}

func (g *Game) DrawGameOver() {

	op := &text.DrawOptions{}

	op.GeoM.Translate(0.5*g.screenWidth-g.Font("small")*float64(len(g.randomQuote))/3.6, 0.4*g.screenHeight)
	text.Draw(g.screen, g.randomQuote, g.GetFontFace("small", true), op)
	g.DrawHiScores()
	g.DrawScores()
}

func (g *Game) DrawHiScores() {
	op := &text.DrawOptions{}
	face := g.GetFontFace("medium", true)
	op.GeoM.Translate(0.1*g.screenWidth, 0.02*g.screenHeight)
	text.Draw(g.screen, fmt.Sprintf("BEST EATING SPREE: %0.0f", g.mostEaten), face, op)
	op.GeoM.Translate(0.6*g.screenWidth, 0)
	text.Draw(g.screen, fmt.Sprintf("HI-SCORE: %0.0f", g.highScore), face, op)
}

func (g *Game) DrawMenu() {

	g.DrawAllNpcFish()

	logoOp, footerOp := &text.DrawOptions{}, &text.DrawOptions{}
	footerOp.GeoM.Translate(0, 0.95*g.screenHeight)
	plainFace := g.GetFontFace("medium", false)
	if !g.menuHidden {
		g.DrawHiScores()
		logoFace := g.GetFontFace("logo", true)
		logoOp.GeoM.Translate(0.3*g.screenWidth, 0.15*g.screenHeight)
		text.Draw(g.screen, title, logoFace, logoOp)

		for i, menuItem := range g.mainMenu {
			menuItem.Draw(g.screen, i == g.activeMenuIndex)
		}

		text.Draw(g.screen, "<", plainFace, footerOp)
		footerOp.GeoM.Translate(0.8*g.screenWidth, 0)
		text.Draw(g.screen, "by Dmitriy Lisovin (2024)", g.GetFontFace("small", false), footerOp)

	} else {
		text.Draw(g.screen, ">", plainFace, footerOp)
	}
}

func (g *Game) DrawOptions() {
	g.DrawAllNpcFish()
	for i, menuItem := range g.optionsMenu {
		menuItem.Draw(g.screen, i == g.activeMenuIndex)
	}
}

func (g *Game) DrawPlanesRecursive(fishes []Fish, count, plane int) {
	if plane < 0 || count <= 0 {
		return
	}
	otherPlaneFishes := []Fish{}
	otherCount := 0
	for i := 0; i < count; i++ {
		switch {
		case int(fishes[i].Plane) == plane:
			fishes[i].Draw()
		case int(fishes[i].Plane) < plane:
			otherPlaneFishes = append(otherPlaneFishes, fishes[i])
			otherCount++
		}
	}
	g.DrawPlanesRecursive(otherPlaneFishes, otherCount, plane-1)
}

func (g *Game) DrawScores() {
	op := &text.DrawOptions{}
	face := g.GetFontFace("medium", true)
	op.GeoM.Translate(0.4*g.screenWidth, 0.5*g.screenHeight)
	text.Draw(g.screen, fmt.Sprintf("FISH EATEN: %0.0f", g.eaten), face, op)
	op.GeoM.Translate(0.03*g.screenWidth, 0.1*g.screenHeight)
	text.Draw(g.screen, fmt.Sprintf("SCORE: %0.0f", g.score), face, op)
}

func (g *Game) DrawVictory() {
	g.DrawAllNpcFish()
	g.playerFish.Draw()
	op := &text.DrawOptions{}
	face := &text.GoTextFace{
		Source: g.fancyFontSource,
		Size:   g.Font("medium"),
	}
	op.GeoM.Translate(0.375*g.screenWidth, 0.2*g.screenHeight)
	text.Draw(g.screen, "CONGRATULATIONS!", face, op)
	op.GeoM.Translate(-0.225*g.screenWidth, 0.1*g.screenHeight)
	text.Draw(g.screen, "You have become the biggest fish in the ocean!", face, op)
	op.GeoM.Translate(0.05*g.screenWidth, 0.1*g.screenHeight)
	text.Draw(g.screen, "Now you can eat anyone with impunity.", face, op)
	g.DrawHiScores()
	g.DrawScores()
}

func (g *Game) End(gameState int) {
	g.gameState = gameState
}

func (g *Game) Font(i string) float64 {
	return g.fontSizes[i]
}

func (g *Game) GameOver() {
	g.End(gameOver)
	g.randomQuote = quotes[rand.Intn(len(quotes))]

}

func (g *Game) GameCycle() error {
	if !g.paused {
		for i, _ := range g.fish {
			g.fish[i].Move()
		}
		g.playerFish.Move()
	}
	if !g.playerFish.Dead {
		g.GetBackgroundColor(g.playerFish.Y)
		if isAnyOfKeysPressed(true, ebiten.KeyP, ebiten.KeyPause, ebiten.KeyEnter) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonCenterRight) {
			g.paused = !g.paused
		}
		if isAnyOfKeysPressed(true, ebiten.KeyEscape) {
			if g.paused {
				g.GoToMenu(false)
			} else {
				g.paused = true
			}
		}
		if g.paused && g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonRightRight) {
			g.GoToMenu(false)
		}
	}
	return nil
}

func (g *Game) GameOverCycle() error {
	if isAnyOfKeysPressed(true, ebiten.KeySpace, ebiten.KeyEnter) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonRightBottom, ebiten.StandardGamepadButtonCenterRight) {
		g.Restart()
	}
	if isAnyOfKeysPressed(true, ebiten.KeyEscape) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonRightRight) {
		g.GoToMenu(true)
	}

	return nil
}

func (g *Game) GenerateFish() {
	g.totalFishCount = int(g.fishPerPlane * g.planeCount)
	g.fish = g.fishStaticArray[0:g.totalFishCount]
	for i := 0; i < g.totalFishCount; i++ {
		g.fish[i].Init(g, getFishType(float64(i), float64(g.totalFishCount)))
	}
}

func (g *Game) GetBackgroundColor(y float64) {
	max := float64(g.screenHeight)
	r, gr, b := getColorComponentByDepth(y, max/3, 128), getColorComponentByDepth(y, max/2, 255), getColorComponentByDepth(y, max, 192)
	g.background = color.RGBA{uint8(r), uint8(gr), uint8(b), 128}
	return
}

func (g *Game) GetFontFace(sizeIndex string, fancy bool) (face *text.GoTextFace) {
	face = &text.GoTextFace{
		Source: g.plainFontSource,
		Size:   g.Font(sizeIndex),
	}
	if fancy {
		face.Source = g.fancyFontSource
	}
	return
}

func (g *Game) GoToMenu(generate bool) {
	g.gameState = gameMenu
	g.menuHidden = false
	g.activeMenuIndex = 0
	if generate {
		g.GenerateFish()
		for i, _ := range g.fish {
			g.fish[i].Randomize()
		}
	}
}

func (g *Game) GoToOptions() {
	g.gameState = gameOptionsMenu
	g.activeMenuIndex = len(g.optionsMenu) - 1

}

func (g *Game) HasMouseMoved() (hasMoved bool) {
	x, y := ebiten.CursorPosition()
	if x != g.prevCurX || y != g.prevCurY {
		hasMoved = true
	}
	g.prevCurX, g.prevCurY = x, y
	return
}

func (g *Game) isAnyGamepadButtonsPressed(just bool, buttons ...ebiten.StandardGamepadButton) (pressed bool) {
	var method func(id ebiten.GamepadID, button ebiten.StandardGamepadButton) bool
	if just {
		method = inpututil.IsStandardGamepadButtonJustPressed
	} else {
		method = ebiten.IsStandardGamepadButtonPressed
	}

	if ebiten.IsStandardGamepadLayoutAvailable(g.gamepadId) {
		for _, button := range buttons {
			if method(g.gamepadId, button) {
				pressed = true
			}
		}

	}
	return
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidthVar, screenHeightVar int) {
	return screenWidth, screenHeight
}

func (g *Game) MenuButtonDown() bool {
	return isAnyOfKeysPressed(true, ebiten.KeyArrowDown, ebiten.KeyS) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonLeftBottom)
}

func (g *Game) MenuButtonLeft() bool {
	return isAnyOfKeysPressed(true, ebiten.KeyA, ebiten.KeyArrowLeft) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonLeftLeft)
}

func (g *Game) MenuButtonRight() bool {
	return isAnyOfKeysPressed(true, ebiten.KeyD, ebiten.KeyArrowRight) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonLeftRight)
}

func (g *Game) MenuButtonUp() bool {
	return isAnyOfKeysPressed(true, ebiten.KeyArrowUp, ebiten.KeyW) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonLeftTop)
}

func (g *Game) MenuCycle() error {
	for i, _ := range g.fish {
		g.fish[i].Move()
	}
	if x, y := ebiten.CursorPosition(); (float64(y) >= 0.9*g.screenHeight && float64(x) < 0.03*g.screenWidth && inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0)) || isAnyOfKeysPressed(true, ebiten.KeyH) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonCenterLeft) {
		g.menuHidden = !g.menuHidden
		return nil
	}
	if g.menuHidden {
		return nil
	}
	if g.MenuButtonDown() {
		g.activeMenuIndex = int(math.Min(float64(g.activeMenuIndex+1), float64(len(g.mainMenu)-1)))
	}
	if g.MenuButtonUp() {
		g.activeMenuIndex = int(math.Max(float64(g.activeMenuIndex-1), 0))
	}
	if g.HasMouseMoved() {
		for i, _ := range g.mainMenu {
			if g.mainMenu[i].DetectHover() {
				g.activeMenuIndex = i
			}
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) || isAnyOfKeysPressed(true, ebiten.KeyEnter, ebiten.KeySpace) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonRightBottom, ebiten.StandardGamepadButtonCenterRight) {
		switch g.activeMenuIndex {
		case 0:
			g.Start()
		case 1:
			g.GoToOptions()
		case 2:
			os.Exit(0)
		}
	}

	return nil
}

func (g *Game) OptionsCycle() error {
	backIndex := len(g.optionsMenu) - 1
	for i, _ := range g.fish {
		g.fish[i].Move()
	}
	if isAnyOfKeysPressed(true, ebiten.KeyEscape) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonRightRight) {
		g.GoToMenu(true)
	}
	if g.MenuButtonDown() {
		g.activeMenuIndex = int(math.Min(float64(g.activeMenuIndex+1), float64(backIndex)))
	}
	if g.MenuButtonUp() {
		g.activeMenuIndex = int(math.Max(float64(g.activeMenuIndex-1), 0))
	}

	if g.HasMouseMoved() {
		for i, _ := range g.optionsMenu {
			if g.optionsMenu[i].DetectHover() {
				g.activeMenuIndex = i
			}
		}
	}
	if g.activeMenuIndex != backIndex {
		if g.MenuButtonRight() {
			if g.optionsMenu[g.activeMenuIndex].ShiftRight() {
				g.ApplyOptions()
			}
		}
		if g.MenuButtonLeft() {
			if g.optionsMenu[g.activeMenuIndex].ShiftLeft() {
				g.ApplyOptions()
			}
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) || isAnyOfKeysPressed(true, ebiten.KeyEnter, ebiten.KeySpace) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonRightBottom) {
		switch {
		case g.activeMenuIndex < backIndex:
			g.optionsMenu[g.activeMenuIndex].CycleRight()
			g.ApplyOptions()
		case g.activeMenuIndex == backIndex:
			g.GoToMenu(false)
		}
	}
	return nil
}

func (g *Game) Restart() {
	g.gameState = gameRunning
	g.score, g.eaten = 0, 0
	for i, _ := range g.fish {
		g.fish[i].Randomize()
	}
	g.playerFish.Reset()
}

func (g *Game) SetDefaultOptions() {
	g.debugEnabled = false
	g.fishReactionsEnabled = true
	g.planeCount = 2
	g.fishPerPlane = 15
	g.fishSizeCap = 45
	g.fishSpeedModifier = 1.0
	g.playerAcceleration = 0.5
	g.playerDeceleration = -0.025
}

func (g *Game) SetFontsSizes() {
	clear(g.fontSizes)
	g.fontSizes["logo"] = 0.15 * g.screenHeight
	g.fontSizes["big"] = 0.1 * g.screenHeight
	g.fontSizes["biggish"] = 0.08 * g.screenHeight
	g.fontSizes["medium"] = 0.05 * g.screenHeight
	g.fontSizes["small"] = 0.03 * g.screenHeight

	return
}

func (g *Game) Start() {
	g.GenerateFish()
	g.Restart()
	g.paused = false
}

func (g *Game) Update() error {
	switch g.gameState {
	case gameRunning:
		return g.GameCycle()
	case gameOver:
		return g.GameOverCycle()
	case gameMenu:
		return g.MenuCycle()
	case gameVictory:
		return g.VictoryCycle()
	case gameOptionsMenu:
		return g.OptionsCycle()
	}
	return nil
}

func (g *Game) UpdateScore(targetSize float64) {
	g.eaten++
	g.score += g.fishSpeedModifier*targetSize*10 + g.fishPerPlane
	g.highScore = math.Max(g.highScore, g.score)
	g.mostEaten = math.Max(g.eaten, g.mostEaten)

}

func (g *Game) VibrateGamepad(milliseconds time.Duration, strong, weak float64) {
	op := &ebiten.VibrateGamepadOptions{
		Duration:        milliseconds * time.Millisecond,
		StrongMagnitude: strong,
		WeakMagnitude:   weak,
	}
	ebiten.VibrateGamepad(g.gamepadId, op)
}

func (g *Game) VibrateGamepadQuick() {
	g.VibrateGamepad(200, 0, 0.5)
}

func (g *Game) VibrateGamepadHeavy() {
	g.VibrateGamepad(500, 0.5, 0)
}

func (g *Game) VictoryCycle() error {
	if isAnyOfKeysPressed(true, ebiten.KeySpace, ebiten.KeyEscape, ebiten.KeyEnter) || g.isAnyGamepadButtonsPressed(true, ebiten.StandardGamepadButtonRightRight, ebiten.StandardGamepadButtonRightBottom, ebiten.StandardGamepadButtonCenterRight) {
		g.GoToMenu(false)
	}
	return nil
}

func (g *Game) Win() {
	for i, _ := range g.fish {
		if g.fish[i].Plane == 0 {
			g.fish[i].SwitchPlane()
		}
	}
	if g.playerFish.Plane == 0 {
		g.playerFish.SwitchPlane()
	}
	g.End(gameVictory)
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(title)
	ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

func getColorComponentByDepth(y, maxDepth, maxColorValue float64) float64 {
	var depth float64
	switch {
	case y < 0:
		depth = 0
	case y > maxDepth:
		depth = maxDepth
	default:
		depth = y
	}
	return maxColorValue * (maxDepth - depth) / maxDepth

}

func preloadImage(img []byte) image.Image {
	buf := bytes.NewBuffer(img)
	m, _, _ := image.Decode(buf)
	return m
}

func loadFont(source []byte) *text.GoTextFaceSource {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(source))
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func isAnyOfKeysPressed(just bool, keys ...ebiten.Key) bool {
	var method func(key ebiten.Key) bool
	if just {
		method = inpututil.IsKeyJustPressed
	} else {
		method = ebiten.IsKeyPressed
	}
	for _, key := range keys {
		if method(key) {
			return true
		}
	}
	return false
}

func NewGame() *Game {
	g := &Game{}
	g.SetDefaultOptions()
	g.screenWidth, g.screenHeight = screenWidth, screenHeight
	g.fontSizes = make(map[string]float64)
	g.SetFontsSizes()
	g.preloadedImages = map[string]image.Image{
		"player":   preloadImage(playerImage),
		"bass":     preloadImage(bassImage),
		"shark":    preloadImage(sharkImage),
		"puffer":   preloadImage(pufferImage),
		"goldfish": preloadImage(goldfishImage),
		"jelly":    preloadImage(jellyImage),
	}
	g.GetBackgroundColor(g.screenHeight / 2)
	g.plainFontSource = loadFont(fixedsys)
	g.fancyFontSource = loadFont(aquawow)
	g.playerFish.Init(g)
	g.CreateMenus()
	g.GoToMenu(true)
	return g
}
