package main

// Extra challenges:
//1. Optimized screen reandering
//   (removed screen.clear and clear only pixel that need to delete use 2D slice of pixel)
//2. Change collor or use collor sceame
//3. Randomeize ball speed and start movement
//4. Use score insted game over
//5. Add in-game bonuses
//6. Dynamicly change ball movement
import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/encoding"
	"github.com/gdamore/tcell/v2"
)

const (
	PaddleSymbol           = tcell.RuneBlock
	BallSymbol             = tcell.RuneDiamond //0x25CD
	InitialBallVelocityRow = 1
	InitialBallVelocityCol = 2
	PaddleHeight           = 4
)

type GameObject struct {
	row, col, width, height int
	velRow, velCol          int
	symbol                  rune
}

var (
	debuglog      string
	isGamePause   bool
	screen        tcell.Screen
	player1Paddle *GameObject
	player2Paddle *GameObject
	ball          *GameObject
	gameObjects   []*GameObject
)

func PrintString(row, col int, text string) {
	for _, r := range text {
		screen.SetContent(col, row, r, nil, tcell.StyleDefault)
		col++
	}
}
func PrintStringCentered(row, col int, str string) {
	col = col - len(str)/2
	PrintString(row, col, str)
}

func PrintChar(row, col, width, height int, ch rune) {
	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			screen.SetContent(col+c, row+r, ch, nil, tcell.StyleDefault)
		}
	}
}

//DrawSatate of Game
func DrawSatate() {
	if isGamePause {
		return
	}
	screen.Clear()
	PrintString(0, 0, debuglog)
	for _, obj := range gameObjects {
		PrintChar(obj.row, obj.col, obj.width, obj.height, obj.symbol)
	}
	screen.Show()
}

func InitScreen() {
	// Initialize screen
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	screen.SetStyle(defStyle)
	screen.Clear()
}

//InitPlayer posision
func InitPlayer() {

	width, height := screen.Size()
	paddleStart := height/2 - PaddleHeight/2

	player1Paddle = &GameObject{
		row: paddleStart, col: 0, width: 1, height: PaddleHeight, symbol: PaddleSymbol,
		velRow: 0, velCol: 0,
	}

	player2Paddle = &GameObject{
		row: paddleStart, col: width - 1, width: 1, height: PaddleHeight, symbol: PaddleSymbol,
		velRow: 0, velCol: 0,
	}

	ball = &GameObject{
		row: 2, col: 2, width: 1, height: 1, symbol: BallSymbol,
		velRow: InitialBallVelocityRow, velCol: InitialBallVelocityCol,
	}

	gameObjects = []*GameObject{
		player1Paddle, player2Paddle, ball,
	}
}

func InitUserInput() chan string {
	inputChan := make(chan string)
	go func() {
		for {
			switch ev := screen.PollEvent().(type) {
			case *tcell.EventKey:
				inputChan <- ev.Name()
			}
		}
	}()
	return inputChan
}

// read input from Background
func ReadInput(inputChan chan string) string {
	var key string
	select {
	case key = <-inputChan:
	default:
		key = ""
	}
	return key
}

func HandlingUserInput(key string) {
	_, screenHeight := screen.Size()
	quit := func() {
		screen.Fini()
		os.Exit(0)
	}
	if key == "Rune[q]" {
		quit()
	} else if key == "Rune[w]" && player1Paddle.row > 0 {
		player1Paddle.row--
	} else if key == "Rune[s]" && player1Paddle.row+player1Paddle.height < screenHeight {
		player1Paddle.row++
	} else if key == "Up" && player2Paddle.row > 0 {
		player2Paddle.row--
	} else if key == "Down" && player2Paddle.row+player2Paddle.height < screenHeight {
		player2Paddle.row++
	} else if key == "Rune[p]" {
		isGamePause = !isGamePause
	}
}

func UpdateState() {
	if isGamePause {
		return
	}
	debuglog = fmt.Sprintf("ball: row:%d, col:%d\npaddle 1: row:%d, col:%d\npaddle 2 : row:%d, col:%d,",
		ball.row, ball.col, player1Paddle.row, player1Paddle.col, player2Paddle.row, player2Paddle.col)

	for i := range gameObjects {
		gameObjects[i].row += gameObjects[i].velRow
		gameObjects[i].col += gameObjects[i].velCol
	}
	if CollidesWithWall(ball) {
		ball.velRow = -ball.velRow
	}
	if CollidesWithPaddle(ball, player1Paddle) || CollidesWithPaddle(ball, player2Paddle) {
		ball.velCol = -ball.velCol
	}
	if CollidesPaddleEdge(ball, player1Paddle) || CollidesPaddleEdge(ball, player2Paddle) {
		ball.velRow = -ball.velRow
		ball.velCol = -ball.velCol
	}
}
func CollidesWithWall(obj *GameObject) bool {
	_, screenHeight := screen.Size()
	return obj.row+obj.velRow < 0 || obj.row+obj.velRow >= screenHeight
}

func CollidesWithPaddle(ball, paddle *GameObject) bool {
	var collidesOnCol bool
	if ball.col < paddle.col {
		collidesOnCol = ball.col+ball.velCol >= paddle.col
	} else {
		collidesOnCol = ball.col+ball.velCol <= paddle.col
	}
	return collidesOnCol &&
		ball.row >= paddle.row &&
		ball.row < paddle.row+PaddleHeight
}

func CollidesPaddleEdge(ball, paddle *GameObject) bool {
	var collidesOnCol bool
	if ball.col < paddle.col {
		collidesOnCol = ball.col+ball.velCol >= paddle.col
	} else {
		collidesOnCol = ball.col+ball.velCol <= paddle.col
	}
	return collidesOnCol && ball.row+ball.velRow == paddle.row ||
		collidesOnCol && ball.row+ball.velRow == paddle.row+PaddleHeight
}
func isGameOver() bool {
	return GetWiner() != ""
}

func GetWiner() string {
	screenWidth, _ := screen.Size()
	if ball.col < 0 {
		return "Player 2"
	} else if ball.col >= screenWidth {
		return "player 1"
	} else {
		return ""
	}
}

func main() {

	encoding.Register()
	InitScreen()
	InitPlayer()
	inputChan := InitUserInput()

	for !isGameOver() {
		// Poll event
		HandlingUserInput(ReadInput(inputChan))
		UpdateState()
		// Update screen
		DrawSatate()

		time.Sleep(75 * time.Millisecond)
	}
	screenWidth, screenHeight := screen.Size()
	winner := GetWiner()
	PrintStringCentered(screenHeight/2-1, screenWidth/2, "Game Over!")
	PrintStringCentered(screenHeight/2, screenWidth/2, fmt.Sprintf("%s Wins...\n", winner))
	screen.Show()
	time.Sleep(3 * time.Second)
	screen.Fini()
}
