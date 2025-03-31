package main

import (
	chess "chess/board"
	"fmt"
	"time"
)

func sleepMilli(delay int) {
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

func main() {
	delay := 1000
	myBoard := chess.NewBoard()
	fmt.Println(myBoard.PrintBoard())
	sleepMilli(delay)
	myBoard.MovePiece(chess.Pawns, chess.NotationToIndex["e2"], chess.NotationToIndex["e4"], chess.Empty)
	fmt.Println(myBoard.PrintBoard())
	sleepMilli(delay)
	myBoard.MovePiece(chess.Pawns, chess.NotationToIndex["e7"], chess.NotationToIndex["e5"], chess.Empty)
	fmt.Println(myBoard.PrintBoard())
	sleepMilli(delay)
	myBoard.MovePiece(chess.Knights, chess.NotationToIndex["g1"], chess.NotationToIndex["f3"], chess.Empty)
	fmt.Println(myBoard.PrintBoard())
	sleepMilli(delay)
	myBoard.MovePiece(chess.Knights, chess.NotationToIndex["b8"], chess.NotationToIndex["c6"], chess.Empty)
	fmt.Println(myBoard.PrintBoard())
	sleepMilli(delay)
	myBoard.MovePiece(chess.Pawns, chess.NotationToIndex["d2"], chess.NotationToIndex["d4"], chess.Empty)
	fmt.Println(myBoard.PrintBoard())
	sleepMilli(delay)
	myBoard.MovePiece(chess.Pawns, chess.NotationToIndex["e5"], chess.NotationToIndex["d4"], chess.Empty)
	fmt.Println(myBoard.PrintBoard())
	sleepMilli(delay)
}
