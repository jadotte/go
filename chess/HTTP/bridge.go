package main

import "C"
import (
	chess "chess/board"
	"encoding/json"
)

var board *chess.Board

//export InitBoard
func InitBoard() {
	board = chess.NewBoard()
}

//export GetBoardTensor
func GetBoardTensor() *C.char {
	tensor := board.ToTensor()
	jsonData, _ := json.Marshal(tensor)
	return C.CString(string(jsonData))
}

//export MakeMove
func MakeMove(start, end C.int, promotion C.int) C.int {
	piece, promo := board.IsLegal(chess.Square(start), chess.Square(end), chess.Piece(promotion))
	if piece == chess.Empty {
		return 0 // Illegal move
	}

	board.MovePiece(piece, chess.Square(start), chess.Square(end), promo)
	board.Turn = board.Turn.Other()
	return 1 // Success
}

func main() {}
