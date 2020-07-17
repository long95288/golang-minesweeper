package main

import (
    "encoding/json"
    "fmt"
    "github.com/therecipe/qt/core"
    "github.com/therecipe/qt/gui"
    "github.com/therecipe/qt/widgets"
    "io/ioutil"
    "math/rand"
    "os"
    "strconv"
    "sync"
    "time"
)
const(
    // 方块类型
    BLOCKTYPE__UNDEFINE int8 = -3 // 未定义
    BLOCKTYPE__MINE int8 = -2 // 地雷
    BLOCKTYPE__FLAG int8 = -1 // 旗子
    BLOCKTYPE__EMPTY int8 = 0 // 空
)
var (
    rows int = 16//行数
    column int = 20  // 列数
    mineNumber int = 20
)
type Mine struct {
    x int
    y int
}

func (m *Mine) setXY(x,y int)  {
    m.x = x
    m.y = y
}
func (m *Mine) getX() int {
    return m.x
}
func (m *Mine) getY() int {
    return m.y
}

type config struct {
    BgImage string `json:"bgImage"`
    LabelImage string `json:"labelImage"`
    AppIcon string `json:"appIcon"`
}
func (c *config) init() {
    data,err := ioutil.ReadFile("conf.json")
    if err != nil {
        c.BgImage = "bg.jpg"
        c.LabelImage = "label.png"
        c.AppIcon="app.png"
        return
    }
    err = json.Unmarshal(data,c)
    if err != nil {
        c.BgImage = "bg.jpg"
        c.LabelImage = "label.png"
        c.AppIcon = "app.png"
    }
}
var (
    app *widgets.QMainWindow
    gameMap [][]int8 // 游戏地图
    flagImage *gui.QPixmap // 旗子图片
    mineImage *gui.QPixmap // 地雷图片
    mineArea *widgets.QWidget // 雷区
    flagNumber int = 10
    mines      []Mine
    timeUsedLCD *widgets.QLCDNumber
    statusLabel   *widgets.QLabel
    flagRemainLCD *widgets.QLCDNumber
    isGameOver bool
    lock sync.Mutex
    bgPixMap *gui.QPixmap
    bgPalette *gui.QPalette
    brush *gui.QBrush
    wHeight int
    wWidth int
    configuration config
)

// 设置初始的游戏地图
func setDefaultMap() {
    for y:= 0;y<rows;y++{
        for x:=0; x <column; x++{
            setGameMapBlock(x,y,BLOCKTYPE__UNDEFINE)
        }
    }
}
// 计数器
func counter() {
    count :=0
    for !isGameOver {
        timeUsedLCD.Display2(count)
        count ++
        time.Sleep(1*time.Second)
    }
}
func startActionHandle(checked bool)   {
    fmt.Println("开始菜单被按下")
    resetMines()
    mineArea.SetEnabled(true)
    // 开始计数
    isGameOver = false
    go counter()
    
}

// 判断(x,y)坐标下面是否是雷
func isMine(x, y int) bool {
    for i:=0;i < mineNumber;i++{
        if mines[i].getX() == x && mines[i].getY() == y{
            // 雷
            return true
        }
    }
    return false
}
// 左键事件处理
func handleLeftButtonPress(event *gui.QMouseEvent) {
    //fmt.Println("按下鼠标左键")
    //fmt.Printf("坐标:x= %d, y= %d \n",event.X()/40,event.Y()/40)
    x,y := event.X()/40,event.Y()/40
    if !isValid(x,y) {
        return
    }
    if isMine(x,y) && gameMap[y][x] == BLOCKTYPE__UNDEFINE{
        // 触雷了
        GameOver()
        return
    }else if gameMap[y][x] == BLOCKTYPE__UNDEFINE {
        // 挖开该点
        surroundMines := getSurroundMines(x, y)
        if surroundMines == 0 {
            // 设置周围区块
            setSurround(x,y)
        }else{
            // 显示雷数
            gameMap[y][x] = int8(surroundMines)
        }
    }else if gameMap[y][x] == BLOCKTYPE__FLAG{
        // 去掉旗帜
        flagNumber ++
        // 设置为未定义
        gameMap[y][x] = BLOCKTYPE__UNDEFINE
    }
    // 判断对局
    if checkWin() {
        WinGame()
    }
    mineArea.Repaint()
}

func GameOver() {
    fmt.Println("游戏结束")
    for i:= 0;i<mineNumber;i++{
        gameMap[mines[i].getY()][mines[i].getX()] = BLOCKTYPE__MINE
    }
    mineArea.Repaint()
    mineArea.SetEnabled(false)
    lock.Lock()
    isGameOver = true
    lock.Unlock()
    
    widgets.QMessageBox_Information(
        mineArea,
        "游戏信息",
        "游戏结束",
        widgets.QMessageBox__Yes,
        widgets.QMessageBox__Yes,
        )
}
func WinGame()  {
    // 赢了游戏
    //fmt.Println("赢了游戏")
    GameOver()
}
func checkWin() bool {
    number := rows * column // 初始化为都没有被揭开过的
    for i := 0;i < rows;i ++{
        for j :=0; j < column;j++{
            if gameMap[i][j] != BLOCKTYPE__UNDEFINE  && gameMap[i][j] != BLOCKTYPE__FLAG{
                // 去掉揭开了的方块,剩下的就是未揭开的和旗子的
                number --
            }
        }
    }
    if number <= mineNumber {
        // 当剩余的方块数小于等于地雷数，便可以判断为游戏结束
        return true
    }
    return false
}
// 获得(x,y)坐标周围的雷的个数
// 分别判断前后左右是否有雷
//
func getSurroundMines(x, y int) int {
   var number,tmpx,tmpy int
   for i := -1; i<= 1;i++{
      for j:= -1;j<= 1;j++{
          tmpx,tmpy = x+i,y+j
          if isValid(tmpx,tmpy) && isMine(tmpx,tmpy) {
              number ++
          }
      }
   }
   return number
}
// 判断x,y坐标是否有效
// x坐标要小于列数
// y坐标小小于行数
func isValid(x, y int) bool {
    if x >= 0 && x < column && y >= 0 && y < rows {
        return true
    }
    return false
}

// 设置(x,y)周围八块地方的雷区显示
//
func setSurround(x, y int) {
    var mineNum,tmpx,tmpy int
    for i:=-1; i <= 1; i++{
        for j:=-1;j <= 1;j++{
            tmpx,tmpy = x+i,y+j
            if isValid(tmpx,tmpy) {
                if gameMap[tmpy][tmpx] == BLOCKTYPE__UNDEFINE {
                    // 没有挖开
                    mineNum = getSurroundMines(tmpx, tmpy)
                    if mineNum > 0 {
                        // 设置该点的值
                        gameMap[tmpy][tmpx] = int8(mineNum)
                    } else if mineNum == 0 {
                        // 递归挖开
                        gameMap[tmpy][tmpx] = BLOCKTYPE__EMPTY
                        setSurround(tmpx, tmpy)
                    }
                }
            }
        }
    }
}
// 鼠标右键
// 如果所点的地方是未挖开的,放旗帜
func handleRightButtonPress(event *gui.QMouseEvent)  {
    //fmt.Println("按下鼠标右键")
    //fmt.Printf("坐标:x= %d, y= %d \n",event.X()/40,event.Y()/40)
    x,y := event.X()/40,event.Y()/40
    if !isValid(x,y) {
        return
    }
    blockType := getGameMapBlock(x,y)
    switch blockType {
    case BLOCKTYPE__UNDEFINE:
        if flagNumber > 0{
            flagNumber -= 1
            setGameMapBlock(x,y,BLOCKTYPE__FLAG)
        }
    }
    mineArea.Repaint()
}

func mousePressHandle(event *gui.QMouseEvent)  {
    if event.Button() == core.Qt__LeftButton {
        handleLeftButtonPress(event)
    }else if event.Button() == core.Qt__RightButton {
        handleRightButtonPress(event)
    }
    event.Accept()
}

// 埋地雷
func setMines() {
    mines = make([]Mine,mineNumber)
    setDefaultMap()
    for i:=0;i<mineNumber;i++{
        for true{
            randomX,randomY := rand.Intn(column),rand.Intn(rows)
            isSeted := false
            for j:=0;j<i;j++{
                if mines[j].getX() == randomX && mines[j].getY() == randomY{
                    isSeted = true
                }
            }
            if !isSeted {
                mines[i].setXY(randomX,randomY)
                break
            }
        }
    }
    //for _,mine := range mines {
    //    fmt.Printf("地雷数据:x = %d;y = %d\n",mine.getX(),mine.getY())
    //}
}

func getGameMapBlock(x,y int) int8  {
    return gameMap[y][x]
}
func setGameMapBlock(x,y int,blockType int8)  {
    gameMap[y][x] = blockType
}

func resetMines() {
    flagNumber = mineNumber
    isGameOver = false
    setMines()
    mineArea.Repaint()
}

// 雷区绘制
func mineAreaPaintHandle(event *gui.QPaintEvent) {
    painter := gui.NewQPainter2(mineArea)
    rowEndPoint := column * 40
    columnEndPoint := rows * 40
    // 绘制行数
    for i := 0;i<= rows;i++{
        painter.DrawLine3(0, 40*i, rowEndPoint, 40*i)
    }
    // 绘制列数
    for i:= 0; i <= column; i++ {
        painter.DrawLine3(40*i,0,40*i,columnEndPoint)
    }
    for i:=0; i<rows; i++{
        for j:=0;j<column;j++{
            blockType := gameMap[i][j]
            switch blockType {
            case BLOCKTYPE__EMPTY:
                //painter.FillRect7(j*40+2,i*40+2,37,37,core.Qt__white)
                painter.DrawText3(j*40+2, i*40+2,"")
            case BLOCKTYPE__FLAG:
                painter.DrawPixmap11(j*40+2,i*40+2,37,37,flagImage)
            case BLOCKTYPE__MINE:
                //fmt.Println("雷区")
                painter.DrawPixmap11(j*40+2,i*40+2,37,37,mineImage)
            case BLOCKTYPE__UNDEFINE:
                // 未定义方框
                painter.FillRect7(j*40+2,i*40+2,37,37,core.Qt__lightGray)
            default:
                // 绘制数字
                painter.SetFont(gui.NewQFont2("黑体",20,400,false))
                painter.DrawText3(j*40+15, i*40-10+40,strconv.Itoa(int(blockType)))
            }
        }
    }
    painter.End()
    flagRemainLCD.Display2(flagNumber)
    event.Accept()
}
// 游戏难度选择
func gameLevelSelect(level int){
    switch level {
    case 1:
        rows = 9
        column = 9
        mineNumber = 10
    case 2:
        rows = 16
        column = 16
        mineNumber = 40
    case 3:
        rows = 16
        column = 30
        mineNumber = 99
    default:
        rows = 9
        column = 9
        mineNumber = 10
    }
    
    // 重新设置map
    gameMap = make([][]int8,rows)
    for i:=0;i<rows;i++{
        gameMap[i] = make([]int8,column)
    }
}
func repaintUI(status int) {
    lock.Lock()
    isGameOver = true
    lock.Unlock()
    timeUsedLCD.Display2(0)
    mineArea.SetMinimumSize2(column*40+2,rows*40+2)
    mineArea.SetMaximumSize2(column*40+2,rows*40+2)
    app.SetFixedSize2(column*40+2+20,rows*40+2+60+60)
    app.Repaint()
}
func InitUI() *widgets.QMainWindow {
    configuration = config{}
    configuration.init()
    gameMap = make([][]int8,rows)
    for i:=0;i<rows;i++{
        gameMap[i] = make([]int8,column)
    }
    app = widgets.NewQMainWindow(nil,0)
    app.SetWindowTitle("扫雷")
    layoutWidget := widgets.NewQWidget(app,core.Qt__Widget)
    app.SetCentralWidget(layoutWidget)
    app.SetWindowIcon(gui.NewQIcon5(configuration.AppIcon))
    
    flagImage = gui.NewQPixmap3("flag.png","",core.Qt__AutoColor)
    mineImage = gui.NewQPixmap3("mine.png","",core.Qt__AutoColor)
    bgPixMap = gui.NewQPixmap3(configuration.BgImage,"",core.Qt__AutoColor)
    bgPalette = gui.NewQPalette()
    brush = gui.NewQBrush()
    setMines()
    
    var gameActions []*widgets.QAction
    // 开始
    startAction := widgets.NewQAction2("开始",app)
    startAction.ConnectTriggered(startActionHandle)
    gameActions = append(gameActions,startAction)
    
    // 退出
    exitAction := widgets.NewQAction3(gui.NewQIcon5("exit.png"),"&退出",app)
    exitAction.SetShortcut(gui.NewQKeySequence2("Ctrl+Q",gui.QKeySequence__NativeText))
    exitAction.SetToolTip("退出游戏")
    exitAction.ConnectTriggered(func(checked bool) {
        app.Close()
    })
    gameActions = append(gameActions,exitAction)
    
    // 应用菜单栏
    menuBar := app.MenuBar()
    gameMenu := menuBar.AddMenu2("游戏")
    gameMenu.AddActions(gameActions)
    
    // 难度选择
    gameMenu = menuBar.AddMenu2("难度")
    var difficultySelection []*widgets.QAction
    var easyLevelAction,mediumLevelAction,hardLevelAction *widgets.QAction
    // 简单难度
    easyLevelAction = widgets.NewQAction2("简单",app)
    easyLevelAction.SetChecked(true)
    easyLevelAction.ConnectTriggered(func(checked bool) {
        easyLevelAction.SetChecked(true)
        mediumLevelAction.SetCheckable(false)
        hardLevelAction.SetCheckable(false)
        gameLevelSelect(1)
        repaintUI(1)
    })
    // 中等难度
    mediumLevelAction = widgets.NewQAction2("中等",app)
    mediumLevelAction.ConnectTriggered(func(checked bool) {
        easyLevelAction.SetChecked(false)
        mediumLevelAction.SetCheckable(true)
        hardLevelAction.SetCheckable(false)
        gameLevelSelect(2)
        repaintUI(2)
    })
    // 困难
    hardLevelAction = widgets.NewQAction2("困难",app)
    hardLevelAction.ConnectTriggered(func(checked bool) {
        fmt.Println("checked:",checked)
        easyLevelAction.SetChecked(false)
        mediumLevelAction.SetCheckable(false)
        hardLevelAction.SetCheckable(true)
        gameLevelSelect(3)
        repaintUI(3)
    })
    difficultySelection = append(difficultySelection,easyLevelAction,mediumLevelAction,hardLevelAction)
    gameMenu.AddActions(difficultySelection)
    // 垂直布局
    VerticalLayout := widgets.NewQVBoxLayout2(layoutWidget)
    // 显示栏布局
    showMessageWidget := widgets.NewQWidget(app,core.Qt__Widget)
    showMessageWidget.SetMaximumHeight(60)
    showMessageWidget.SetMinimumHeight(60)
    showMessageLayout := widgets.NewQHBoxLayout2(showMessageWidget)
    // 已用时间
    timeUsedLCD = widgets.NewQLCDNumber(showMessageWidget)
    timeUsedLCD.SetObjectName("timeUsedLCD")
    timeUsedLCD.SetDecMode()
    // 当前状态
    statusLabel = widgets.NewQLabel(showMessageWidget,0)
    statusLabel.SetObjectName("statusLabel")
    statusLabel.SetMaximumWidth(60)
    
    // 剩余旗子
    flagRemainLCD = widgets.NewQLCDNumber(showMessageWidget)
    flagRemainLCD.SetObjectName("flagRemainLCD")
    flagRemainLCD.SetDecMode()
    showMessageLayout.AddWidget(timeUsedLCD,0,0)
    showMessageLayout.AddWidget(statusLabel,0,0)
    showMessageLayout.AddWidget(flagRemainLCD,0,0)
    VerticalLayout.AddWidget(showMessageWidget,0,0)
    
    // 雷区
    mineArea = widgets.NewQWidget(app,core.Qt__Widget)
    mineArea.SetMinimumSize2(column*40+2,rows*40+2)
    mineArea.SetMaximumSize2(column*40+2,rows*40+2)
    mineArea.SetObjectName("mineArea")
    mineArea.SetEnabled(false)
    
    VerticalLayout.AddWidget(mineArea,0,0)
    
    mineArea.ConnectPaintEvent(mineAreaPaintHandle)
    mineArea.ConnectMousePressEvent(mousePressHandle)
    setStyle(app)
    return app
}

func setStyle(app *widgets.QMainWindow) {
    labelBg := gui.NewQPixmap3(configuration.LabelImage,"",core.Qt__AutoColor).Scaled2(60,60,core.Qt__IgnoreAspectRatio,core.Qt__SmoothTransformation)
    statusLabel.SetPixmap(labelBg)
    style, err := ioutil.ReadFile("style.qss")
    if err != nil {
        return
    }
    app.SetStyleSheet(string(style))
    app.ConnectPaintEvent(func(event *gui.QPaintEvent) {
        // 重设图片宽高以适应应用大小
        if !(wHeight == app.Height() && wWidth == app.Width()) {
            bgPixMapTmp := bgPixMap.Scaled2(app.Width(),app.Height(),core.Qt__IgnoreAspectRatio,core.Qt__SmoothTransformation)
            brush.SetTexture(bgPixMapTmp)
            bgPalette.SetBrush(gui.QPalette__Background,brush)
            app.SetPalette(bgPalette)
            wHeight = app.Height()
            wWidth = app.Width()
        }
        event.Accept()
    })
}
func main() {
    // 创建应用程序
    widgets.NewQApplication(len(os.Args),os.Args)
    app := InitUI()
    app.Show()
    widgets.QApplication_Exec()
}