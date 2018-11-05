package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	originLines []string
	targetLines []string
	resultLines chan []string
	eol         string
	syc         sync.WaitGroup
	execNum     = 0
	originLen   = 0
	targetLen   = 0
	mutex       sync.Mutex     //写文件锁
	pregDone    chan bool      //匹配处理信号
	targetIndex map[string]int //索引
)

//init函数 初始化 每次启动的时候 这个函数会被调用
func init() {
	//判断当前运行环境
	if runtime.GOOS == "windows" {
		eol = "\r\n"
	} else {
		eol = "\n"
	}
	//初始化两个通道
	resultLines = make(chan []string)
	pregDone = make(chan bool)
	targetIndex = make(map[string]int)

}

//读取源文件
func readFile(filename string, arrTarget int) bool {
	//打开文件
	originFile, err := os.Open(filename)
	if err != nil {
		log.Fatal("读取文件 " + filename + " 错误" + err.Error())
	}
	//defer 在函数结束时 defer注册的函数会被调用
	defer originFile.Close()
	//读取文件内容
	reader := bufio.NewReader(originFile)
	for {
		//一行一行的读
		line, _, err := reader.ReadLine()
		switch arrTarget {
		case 1:
			originLines = append(originLines, string(line))
		case 2:
			targetLines = append(targetLines, string(line))
		}

		if err != nil {
			fmt.Println("文件 " + filename + " 读取完毕...")
			return true
		}
	}
	return false
}

//给目标文件内容做一个索引
func makeIndexForTarget() {
	for k, tl := range targetLines {
		tdata := strings.Replace(tl, "\"", "", -1)
		tas := strings.Split(tdata, ",")
		if len(tas) > 20 && tas[20] != "" {
			targetIndex[getIndexKey(tas[7], tas[11])] = k
		}

	}
}

func getIndexKey(key1, key2 string) string {
	return "k1_" + key1 + "k2_" + key2
}

//匹配数据
func matchResultData() {

	originLen = len(originLines)
	targetLen = len(targetLines)
	fmt.Println("原始数据长度:" + strconv.Itoa(originLen))
	fmt.Println("目标数据长度:" + strconv.Itoa(targetLen))
	//计算数据分割点
	checkLen := int(math.Floor(float64(originLen / 4))) //2
	checkPoint1 := checkLen                             //2
	checkPoint2 := checkPoint1 + checkLen               //4
	checkPoint3 := checkPoint2 + checkLen               //6
	//	将元数据分割为4个组 由4个携程同时处理
	go runtimeMatch(originLines[:checkLen])
	go runtimeMatch(originLines[checkPoint1:checkPoint2])
	go runtimeMatch(originLines[checkPoint2:checkPoint3])
	go runtimeMatch(originLines[checkPoint3:])

	<-pregDone
	fmt.Println()
	fmt.Println("文件匹配完毕...")
}

//开启一个携程 打印工作进度
func doPrintProgress() {
	mutex.Lock()
	execNum++
	mutex.Unlock()
	fmt.Printf("\r进度:%d/%d", execNum, originLen)
	if execNum >= originLen {
		fmt.Println("结果文件写入完毕")
		pregDone <- true
	}

}

//做匹配工作的携程
func runtimeMatch(Lines []string) {
	var oas []string
	var resArr []string

	for _, ol := range Lines {
		odata := strings.Replace(ol, "\"", "", -1)
		oas = strings.Split(odata, ",")
		resArr = oas
		go doPrintProgress()
		if len([]byte(odata)) < 1 {
			continue
		}
		key := getIndexKey(oas[7], oas[11])
		if _, ok := targetIndex[key]; ok {
			resArr = append(resArr, "1")
		}
		if len(resArr) > 0 {
			syc.Add(1)
			resultLines <- resArr
		}
	}
}

//一个写文件的携程 运行之后 监控通道内的数据 如果有数据被
//传到通道内 会取出来写到文件里
func writeResultFile(filename string) {

	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic("创建结果文件失败:" + err.Error())
	}
	for {
		select {
		case line := <-resultLines:
			tmp := ""
			for _, v := range line {
				tmp += "\"" + v + "\","
			}
			tmp = strings.TrimRight(tmp, ",") + eol
			io.WriteString(file, tmp)
			syc.Done()
		}
	}

}

func main() {
	originFilename := flag.String("origin", "file1.csv", "原始文件的地址")
	targetFilename := flag.String("target", "file2.csv", "存有更新数据的文件")
	resultFilename := flag.String("result", "result.csv", "对比结果文件名称")
	flag.Parse()
	fmt.Println("导出文件名称:" + *originFilename)
	fmt.Println("历史文件名称:" + *targetFilename)
	fmt.Println("结果文件名称:" + *resultFilename)

	readFile(*originFilename, 1)

	targetFiles := strings.Split(*targetFilename, ",")
	for _, targetFile := range targetFiles {
		readFile(targetFile, 2)
	}
	makeIndexForTarget()

	go writeResultFile(*resultFilename)
	fmt.Println("开始文件匹配...")
	matchResultData()

	syc.Wait()

}
