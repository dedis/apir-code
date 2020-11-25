package main

import (
	"errors"
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	db := CreateAsciiDatabase()
	result := ""
	c := Client{}
	s0 := Server{}
	s1 := Server{}
	for i := 0; i < 136; i++ {
		startQuery := time.Now()
		q0, q1, st := c.Query(i)
		elapsedQuery := time.Since(startQuery)
		fmt.Printf("Query took %s \n", elapsedQuery)

		startAnswerFirst := time.Now()
		a0 := s0.Answer(db, q0)
		elapsedAnswerFirst := time.Since(startAnswerFirst)
		fmt.Printf("Answer first took %s \n", elapsedAnswerFirst)

		startAnswerSecond := time.Now()
		a1 := s1.Answer(db, q1)
		elapsedAnswerSecond := time.Since(startAnswerSecond)
		fmt.Printf("Answer second took %s \n", elapsedAnswerSecond)

		startReconstruct := time.Now()
		x, err := c.Reconstruct(a0, a1, st)
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
