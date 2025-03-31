package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	chess "chess/board"
)

func playTUI() {
	var board *chess.Board = chess.NewBoard()
	counter := 0
	for counter < 50 {
		fmt.Println(board.PrintBoard())
		fmt.Printf("Turn: %v (0=White, 1=Black)\n", board.Turn)
		if !requestMove(board, &counter) {
			continue
		}
		counter++
		board.Turn = board.Turn.Other()
	}
	fmt.Println("Draw by fifty moves!")
}

var pieceMap = map[string]chess.Piece{
	"Pawn": chess.Pawns, "Pawns": chess.Pawns, "pawn": chess.Pawns, "pawns": chess.Pawns,
	"Knight": chess.Knights, "Knights": chess.Knights, "knight": chess.Knights, "knights": chess.Knights,
	"Bishop": chess.Bishops, "Bishops": chess.Bishops, "bishop": chess.Bishops, "bishops": chess.Bishops,
	"Rook": chess.Rooks, "Rooks": chess.Rooks, "rook": chess.Rooks, "rooks": chess.Rooks,
	"Queen": chess.Queens, "Queens": chess.Queens, "queen": chess.Queens, "queens": chess.Queens,
	"King": chess.Kings, "Kings": chess.Kings, "king": chess.Kings, "kings": chess.Kings,
}

func requestMove(board *chess.Board, counter *int) bool {
	fmt.Println("Please input move.")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return false
	}
	var start chess.Square
	var end chess.Square
	var promotion chess.Piece
	var piece chess.Piece
	input = strings.TrimSpace(input)
	move := strings.Split(input, " ")
	if input == "draw" || input == "Draw" {
		return true
	}
	if len(move) < 2 || len(move) > 3 {
		fmt.Println("Please provide at two words if not promoting, and three words if promoting.")
		return false
	}
	if len(move) == 3 {
		promotion = pieceMap[move[2]]
	} else {
		promotion = chess.Empty
	}
	start = chess.NotationToIndex[move[0]]
	end = chess.NotationToIndex[move[1]]
	piece, promotion = board.IsLegal(start, end, promotion)
	fmt.Println(start)
	if piece == chess.Empty {
		fmt.Println("Move not legal.")
		return false
	}
	if piece == chess.Pawns {
		*counter = 0
	}
	if board.MovePiece(piece, start, end, promotion) {
		*counter = 0
	}
	return true
}

func main() {
	playTUI()
}
