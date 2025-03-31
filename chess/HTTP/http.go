package main

import (
	chess "chess/board"
	"encoding/json"
	"fmt"
	"net/http"
)

type MoveRequest struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	Promotion string `json:"promotion,omitempty"`
}

type BoardResponse struct {
	FEN         string            `json:"fen"`
	LegalMoves  []string          `json:"legal_moves"`
	IsTerminal  bool              `json:"is_terminal"`
	GameResult  int               `json:"winner"` // 1 for white win, -1 for black win, 0 for draw
	BoardTensor [8][8][19]float32 `json:"board_tensor"`
}

func main() {
	board := chess.NewBoard()

	// Endpoint to get the current board state
	http.HandleFunc("/board", func(w http.ResponseWriter, r *http.Request) {
		response := BoardResponse{
			BoardTensor: board.ToTensor(),
			// Add other fields...
		}
		json.NewEncoder(w).Encode(response)
	})

	// Endpoint to make a move
	http.HandleFunc("/move", func(w http.ResponseWriter, r *http.Request) {
		var moveReq MoveRequest
		err := json.NewDecoder(r.Body).Decode(&moveReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		start := chess.NotationToIndex[moveReq.Start]
		end := chess.NotationToIndex[moveReq.End]

		// Handle promotion logic
		var promotion chess.Piece = chess.Empty
		if moveReq.Promotion != "" {
			// Map promotion string to piece
		}

		piece, promo := board.IsLegal(start, end, promotion)
		if piece == chess.Empty {
			http.Error(w, "Illegal move", http.StatusBadRequest)
			return
		}

		board.MovePiece(piece, start, end, promo)
		board.Turn = board.Turn.Other()

		// Return updated board
		response := BoardResponse{
			BoardTensor: board.ToTensor(),
			// Add other fields...
		}
		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("Chess engine server running on :8080")
	http.ListenAndServe(":8080", nil)
}
