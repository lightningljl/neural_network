package main
import(
    "flag"
    "fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"time"
	"strconv"
	"encoding/json"
	"math/rand"
	"github.com/garyburd/redigo/redis"
	"github.com/satori/go.uuid"
	"strings"
)
//服务器地址
var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
//redis连接
var redisConn redis.Conn
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var group map[Client]bool

//每个用户的结构体
type Client struct {
    userId string
	conn *websocket.Conn
}

//房间信息
type House struct {
	HouseId string
	User []string
	Mj Majiang
	HandsBrandList map[string]HandsBrand
}

//返回消息的格式
type Message struct {
	Result int
	FunctionId int
	Data interface{}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	userId := strconv.FormatInt(time.Now().Unix(), 10)
	newClient := Client{ userId,  c}
	group[newClient] = true

	
	var responseMessage []byte
	
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("获取前段的数据:", err)
			break
		}
		request := strings.Split(string(message), "|||")
        if err != nil {
        	log.Println("解码出现错误:", err)
			break
        }
		switch request[0] {
			//创建房间
		    case "1":
			    responseMessage = CreateHouse(userId, request[1])
		        err = c.WriteMessage(mt, responseMessage)
		    //加入房间
		    case "2":
			    responseMessage = EnterHouse(userId, request[1])
		        err = c.WriteMessage(mt, responseMessage)
            //发牌
            case "3":
                
			case "broad":
			    for key, _ := range group {
				    res := "this is broadcast message "+userId
					w, _ := key.conn.NextWriter(websocket.TextMessage)
					w.Write([]byte(res))
				}
		    default:
			    res := "ni hao "+userId
		        err = c.WriteMessage(mt, []byte(res)) 
		}
		
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

/**
 * 创建房间方法
 */ 
func CreateHouse( userId string, content string ) ([]byte) {
    message := Message{Result:0, FunctionId:1, Data:"创建放假数据结构出错"}
	var mj Majiang
    //将发送的内容转换为结构体
    err := json.Unmarshal([]byte(content), &mj)
    if err != nil {
    	log.Println("创建房间数据结构出错:", err)
    	return FormatResult(message)
    }
    //初始化房间信息
    house := House{HouseId:uuid.NewV4().String(), Mj:mj}
    house.User = append(house.User, userId)
    
    //将数据存储如redis
    store := Store(house.HouseId, house)
    if !store {
        message.Data = "房间信息存储redis失败"
    	return FormatResult(message)
    }
    _, errRedis := redisConn.Do("set", "user_"+userId, house.HouseId)
    if errRedis != nil {
        log.Println("存储redis用户信息出错:", errRedis)
        message.Data = "存储redis用户信息出错"
        return FormatResult(message)
    }
    message = Message{Result:1, FunctionId:1, Data:house.HouseId}
    return FormatResult(message)
}

/**
 * 加入房间方法
 */ 
func EnterHouse( userId string, houseId string ) ([]byte) {
    message := Message{Result:0, FunctionId:2, Data:"获取用户信息失败"}
	//检查用户是否在房间
	key := "user_"+userId
	user, errRedis := redisConn.Do("get", key)
    if errRedis != nil {
        log.Println("获取redis用户信息出错:", errRedis)
        return FormatResult(message)
    }
    //判断用户是否已经在这个房间，并且房间存在
    if  user != "0" && user != nil {
    	//TODO
    }
    //判断放假是否存在或者人数是否已满
	houseInfo, result := HouseInfo(houseId)
    if result == false {
        message.Data = "房间信息不存在"
        return FormatResult(message)
    }
	//判断人数等信息
    if len(houseInfo.User) >= houseInfo.Mj.PoepleNumber {
        message.Data = "房间人数已满"
        return FormatResult(message)
    }
    //正常的加入房间
    houseInfo.User = append(houseInfo.User, userId)
    //将数据存储如redis
    store := Store(houseInfo.HouseId, houseInfo)
    if !store {
        message.Data = "房间存储失败"
        return FormatResult(message)
    }
    _, errRedis = redisConn.Do("set", key, houseId)
    if errRedis != nil {
        log.Println("存储redis用户信息出错:", errRedis)
        message.Data = "存储redis用户信息出错"
        return FormatResult(message)
    }
    message.Result = 1
    message.Data = "房间加入成功"
    return FormatResult(message)
}

/**
 * 将数据存入redis，返回存入成功或者失败
 */ 
func Store( key string, data interface{} ) bool {
	//先将数据转为json
	jsonData ,err := json.Marshal( data )
    if err != nil {
        fmt.Println("redis存储时, json转换失败")
    }
    //将转换好的数据存储如redis
    _, errRedis := redisConn.Do("set", key, jsonData)
    if errRedis != nil {
        log.Println("将数据存入redis失败:", errRedis)
        return false
    }
    return true
}

//将结果格式话为前段
func FormatResult(message Message) ([]byte) {
    //将房间信息转为json
    resultMessage,errTransfer := json.Marshal( message )
    if errTransfer != nil {
        log.Println("转换json信息失败:", errTransfer)
        return []byte(`{"Result":0,"Function":0,"Data":"json\u8f6c\u6362\u51fa\u9519"}`)
    }
    return resultMessage
}

/**
 * 将房间信息获取出来
 */ 
func HouseInfo(houseId string) (House, bool) {
    //将信息反redis
    var houseInfo House
    houseRedis, errRedis := redisConn.Do("get", houseId)
    if errRedis != nil {
        log.Println("获取redis房间信息出错:", errRedis)
        return houseInfo,false
    }
     //将发送的内容转换为结构体
    err := json.Unmarshal([]byte(houseRedis.([]uint8)), &houseInfo)
    if err != nil {
    	log.Println("解析房间信息出错:", err)
    	return houseInfo,false
    }
    return houseInfo,true
}


func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello single!\n")
}


func main() {
	//服务地址
	flag.Parse()
    //redis链接
    redisConn, _ = redis.Dial("tcp", "127.0.0.1:6379")
    defer redisConn.Close()

    //初始化组
    group = make(map[Client]bool)

	http.HandleFunc("/", echo)
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/mj", mj)
	log.Fatal(http.ListenAndServe(*addr, nil))
}


func mj(w http.ResponseWriter, r *http.Request) {
	aa,err := json.Marshal( brand )
	if err != nil {
        fmt.Fprintf(w, string(aa))
    }
    mj := Majiang{4,1}
    handsBrandList := mj.initHandsBrand()
    aa,err = json.Marshal( handsBrandList )
    if err != nil {
        fmt.Println(err.Error())
    }
	fmt.Fprintf(w, string(aa))
}




//麻将部分
type Majiang struct {
	PoepleNumber int    //参与人数
	Dyj int    //带幺九
}

//用户手牌数据结构
type HandsBrand struct {
	UserId int
	Brand [3][9]int
	NetBrand [][2]int
}

//麻将一维数据结构，有顺序，代表筒，条，万
var brand = [3][9]int{ 
	{4,4,4,4,4,4,4,4,4},
	{4,4,4,4,4,4,4,4,4},
	{4,4,4,4,4,4,4,4,4},
} 
//初始化一维的空的手牌
var brandEmpty = [3][9]int{ 
	{0,0,0,0,0,0,0,0,0},
	{0,0,0,0,0,0,0,0,0},
	{0,0,0,0,0,0,0,0,0},
} 

var brandShuffle [][2]int

//初始手牌化方法
func (mj *Majiang) initHandsBrand() ( map[string]HandsBrand ) {
	//将手牌转为1维
	for i := 0; i < 3; i++ {
		for j := 0; j < 9; j++ {
			for k := 0; k < 4; k++ {
				brandShuffle = append(brandShuffle, [2]int{i,j})
			}
		}
	}
	//将手牌打乱，模拟洗牌
	//根据时间设置随机数种子
    rand.Seed(int64(time.Now().Nanosecond()))
    for i := 0; i < 108; i++ {
    	newKey := rand.Intn(107)
    	temp   := brandShuffle[i]
    	brandShuffle[i] = brandShuffle[newKey]
    	brandShuffle[newKey] = temp
    }
    //根据人来初始化手牌
    userHandBrandList := make(map[string]HandsBrand)
    start := 0
    for i := 0; i < mj.PoepleNumber; i++ {
    	//每个用户截取13张牌
        brandTemp := brandShuffle[start:13+start]
        //将二维的牌结构转换为一维的
        handsBrand :=  brandEmpty
        for _, value := range brandTemp {
        	color  := value[0]
        	number := value[1]
        	handsBrand[color][number] = handsBrand[color][number] + 1
        }
    	userHandBrandList[strconv.Itoa(i)] = HandsBrand{i, handsBrand, brandTemp}
    	start = start + 13
    }
    //将当前桌面上的牌一维清理
    brandShuffle = brandShuffle[start:]
    return userHandBrandList
}

