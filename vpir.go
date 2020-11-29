package main

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

func main() {
	start := time.Now()
	db := CreateAsciiDatabase()
	result := ""
	c := Client{}
	s0 := Server{}
	s1 := Server{}
	s2 := Server{}
	for i := 0; i < 136; i++ {
		startQuery := time.Now()
		queries, st := c.Query(i)
		elapsedQuery := time.Since(startQuery)
		fmt.Printf("Query took %s \n", elapsedQuery)

		startAnswerFirst := time.Now()
		a0 := s0.Answer(db, queries[0])
		elapsedAnswerFirst := time.Since(startAnswerFirst)
		fmt.Printf("Answer first took %s \n", elapsedAnswerFirst)

		startAnswerSecond := time.Now()
		a1 := s1.Answer(db, queries[1])
		elapsedAnswerSecond := time.Since(startAnswerSecond)
		fmt.Printf("Answer second took %s \n", elapsedAnswerSecond)

		startAnswerThird := time.Now()
		a2 := s2.Answer(db, queries[2])
		elapsedAnswerThird := time.Since(startAnswerThird)
		fmt.Printf("Answer third took %s \n", elapsedAnswerThird)

		answers := make([]*big.Int, 3)
		answers[0] = a0
		answers[1] = a1
		answers[2] = a2

		startReconstruct := time.Now()
		x, err := c.Reconstruct(answers, st)
		elapsedReconstruct := time.Since(startReconstruct)
		fmt.Printf("Reconstruct took %s \n", elapsedReconstruct)
		if err != nil {
			panic(err)
		}
		result += x.String()
	}
	b, err := bitStringToBytes(result)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	elapsed := time.Since(start)
	fmt.Printf("Took %s", elapsed)
}

func bitStringToBytes(s string) ([]byte, error) {
	b := make([]byte, (len(s)+(8-1))/8)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '1' {
			return nil, errors.New("not a bit")
		}
		b[i>>3] |= (c - '0') << uint(7-i&7)
	}
	return b, nil
}
