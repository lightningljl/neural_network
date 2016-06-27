package main

import(
    "fmt"
    "math"
    "math/rand"
    "time"
)

//const first,second,third,alpha = 784,100,10,0.35 
const first,second,third,alpha = 10,5,10,0.35 

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
	fmt.Println(weight1)
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
	
}
