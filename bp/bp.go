package main

import(
    //"fmt"
    "math"
    "math/rand"
    "time"
    "os"
    "image/color"
    "image/png"  
    "image"  
    "strconv"  
)

const first,second,third,alpha = 784,100,10,0.35 
//const first,second,third,alpha = 10,5,10,0.35 

var input [first]int
var target [third]int
var weight1[first][second]float64
var weight2[second][third]float64
var output1[second]float64
var output2[third]float64
var delta1[second]float64
var delta2[third]float64
var b1[second]float64
var b2[third]float64

func main() {
    initialWeight()
	//fmt.Println(weight1)
	training()
}

//激活函数
func segemod( x float64 ) float64 {
	return 1.0 / (1.0 + math.Exp(-x));
}

//初始化权重
func initialWeight() {
	randSeed := rand.New(rand.NewSource(time.Now().UnixNano()))
	//初始化weight1
	for i := 0; i < first; i++ {
		for j := 0; j < second; j++ {
			weight1[i][j] = randSeed.Float64()-0.5
		}
	}
	//初始化weight2
	for i := 0; i < second; i++ {
		for j := 0; j < third; j++ {
			weight2[i][j] = randSeed.Float64()-0.5
		}
	}
	//初始化偏执
	for i := 0; i < second; i++ {
		b1[i] = randSeed.Float64()-0.5
	}
	for i := 0; i < third; i++ {
		b2[i] = randSeed.Float64()-0.5
	}
}

//训练
func training() {
	//读取训练文件
	testFile, _ := os.Open("t10k-images.idx3-ubyte")
	labelFile,_ := os.Open("train-labels.idx1-ubyte")
	defer testFile.Close()
	imageBuffer := make([]byte, first)
	labelBuffer := make([]byte, third)
	number := 1
    for{
        n, _ := testFile.Read(imageBuffer)
        if 0 == n { 
        	break
        }
        m, _ := labelFile.Read(labelBuffer)
        //bufferToImage( imageBuffer, strconv.Itoa( number ) )
        for i := 0; i < first; i++){
	        if imageBuffer[i] < 128 {
	        	input[i] = 0
	        } else {
	        	input[i] = 1
	        }
		}
		key := labelBuffer[0]
		for j := 0; j < third; j++ {
			target[j] = 0
		}
        target[key] = 1

        number = number + 1
    }
}

//第一层训练
func firstFloorTraining( ) {
	for i := 0; i < second; i++ {
		sigma := 0.0
		for j := 0; j < first; j++ {
			sigma = sigma + input[j] * weight1[j][i];
		}
		sigma = sigma + b1[i]
		output1[i] = segemod(sigma)
	}
}

//第二层训练
func secondFloorTraining() {
	
}

func bufferToImage( imageBuffer []byte, imageName string ) {
	dx,dy := 28,28
	gray := image.NewGray(image.Rect(0, 0, dx, dy)) 
	number := 0 
	for y := 0; y < dy; y++ {
		for x := 14; x < dx; x++ {
			gray.Set(x, y, color.Gray{imageBuffer[number]})
			number = number + 1
		}
		for x := 0; x < 14; x++ {
			gray.Set(x, y, color.Gray{imageBuffer[number]})
			number = number + 1
		}
	}

	newPath := "image/"+imageName+".png"  
    newFile, _ := os.Create(newPath)  
    png.Encode(newFile, gray)  
}
