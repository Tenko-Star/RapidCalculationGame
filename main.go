package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Quiz struct {
	question string
	answer   int
}

var quizChan = make(chan Quiz, 1)
var isStop = make(chan bool, 1)
var wg = &sync.WaitGroup{}
var symbolMap = map[int]string{
	0: "+",
	1: "-",
	2: "*",
	3: "/",
}

var (
	createIsDone = make(chan bool)
	printIsDone  = make(chan bool)
)

var score int

func main() {
	go createQuiz()
	go printQuiz()

	<-createIsDone
	<-printIsDone

	wg.Wait()
	fmt.Println("Score:", score)
}

func createQuiz() {
	wg.Add(1)
	defer wg.Done()

	// set rand seed
	rand.Seed(time.Now().UnixNano())

	createIsDone <- true

	for {
		var (
			question string
			answer   int
		)

		symbol, origin := getRandomSymbol()

		var (
			numA int
			numB int
		)

		switch origin {
		case 0:
			fallthrough
		case 1:
			numA = rand.Intn(100)
			numB = rand.Intn(100)
		case 2:
			fallthrough
		case 3:
			numA = rand.Intn(10)
			numB = rand.Intn(10)
		}

		question = fmt.Sprintf("%d %s %d = ?", numA, symbol, numB)
		answer = calc(numA, numB, origin)

		select {
		case <-isStop:
			return
		case quizChan <- Quiz{question, answer}:
			continue
		}
	}
}

func calc(numA int, numB int, symbol int) int {
	switch symbol {
	case 0:
		return numA + numB
	case 1:
		return numA - numB
	case 2:
		return numA * numB
	case 3:
		return numA / numB
	default:
		return 0
	}
}

func printQuiz() {
	wg.Add(1)
	defer wg.Done()

	printIsDone <- true

	for {
		var quiz Quiz

		select {
		case quiz = <-quizChan:
			fmt.Println(quiz.question)
			answer, stop := scan()
			if stop == true {
				isStop <- true
				return
			}

			if answer == quiz.answer {
				fmt.Println("Correct!")
				score++
			} else {
				fmt.Println("Wrong!")
			}
		}
	}
}

func scan() (int, bool) {
	var buf string

	timeout := time.NewTimer(time.Second * 5)
	select {
	case <-timeout.C:
		fmt.Println("Timeout!")
		return 0, true
	case buf = <-getInputChan():
		break
	}

	if buf == "stop" || buf == "" {
		return 0, true
	}

	result, err := strconv.Atoi(buf)
	if err != nil {
		fmt.Println(err)
		return 0, true
	}

	return result, false
}

func getRandomSymbol() (string, int) {
	value := rand.Intn(3)
	return symbolMap[value], value
}

func getInputChan() <-chan string {
	ch := make(chan string)
	go func() {
		var buf string
		fmt.Scanln(&buf)
		ch <- buf
	}()
	return ch
}
