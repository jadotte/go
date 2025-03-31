// In chess/api/server.go
package main

import (
	chess "chess/board"
	"encoding/json"
	"fmt"
	"log"
	"math/bits"
	"net/http"
	"strconv"
)

type MoveRequest struct {
	Start     string `json:"start"`     // Chess notation (e.g., "e2")
	End       string `json:"end"`       // Chess notation (e.g., "e4")
	Promotion string `json:"promotion"` // Optional promotion piece
}

type BoardResponse struct {
	Board       string            `json:"board"`       // Pretty-printed board
	Turn        string            `json:"turn"`        // "White" or "Black"
	LegalMoves  []Move            `json:"legalMoves"`  // All legal moves
	IsCheck     bool              `json:"isCheck"`     // Whether current player is in check
	BoardTensor [8][8][19]float32 `json:"boardTensor"` // Neural network input representation
}

var pieceMap = map[string]chess.Piece{
	"knight": chess.Knights, "n": chess.Knights,
	"bishop": chess.Bishops, "b": chess.Bishops,
	"rook": chess.Rooks, "r": chess.Rooks,
	"queen": chess.Queens, "q": chess.Queens,
}

// Global board instance
var board *chess.Board

func boardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Return current board state
		response := prepareBoardResponse()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var moveReq MoveRequest
	err := json.NewDecoder(r.Body).Decode(&moveReq)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the move format
	start, ok := chess.NotationToIndex[moveReq.Start]
	if !ok {
		http.Error(w, "Invalid start position: "+moveReq.Start, http.StatusBadRequest)
		return
	}

	end, ok := chess.NotationToIndex[moveReq.End]
	if !ok {
		http.Error(w, "Invalid end position: "+moveReq.End, http.StatusBadRequest)
		return
	}

	// Handle promotion
	var promotion chess.Piece = chess.Empty
	if moveReq.Promotion != "" {
		promotion, ok = pieceMap[moveReq.Promotion]
		if !ok {
			http.Error(w, "Invalid promotion piece: "+moveReq.Promotion, http.StatusBadRequest)
			return
		}
	}

	// Check if the move is legal
	piece, finalPromotion := board.IsLegal(start, end, promotion)
	if piece == chess.Empty {
		http.Error(w, "Illegal move", http.StatusBadRequest)
		return
	}

	// Make the move
	board.MovePiece(piece, start, end, finalPromotion)
	board.Turn = board.Turn.Other()

	// Return the updated board state
	response := prepareBoardResponse()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Reset the board to starting position
	board = chess.NewBoard()

	response := prepareBoardResponse()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func prepareBoardResponse() BoardResponse {
	return BoardResponse{
		Board:       board.PrintBoard(),
		Turn:        getTurnString(),
		LegalMoves:  getLegalMoves(),
		IsCheck:     isInCheck(),
		BoardTensor: board.ToTensor(),
	}
}

func getTurnString() string {
	if board.Turn == chess.White {
		return "White"
	}
	return "Black"
}

type Move struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	Promotion string `json:"promotion,omitempty"`
}

// Add this to convert square indices to algebraic notation
var IndexToNotation = map[chess.Square]string{
	0: "a1", 1: "b1", 2: "c1", 3: "d1", 4: "e1", 5: "f1", 6: "g1", 7: "h1",
	8: "a2", 9: "b2", 10: "c2", 11: "d2", 12: "e2", 13: "f2", 14: "g2", 15: "h2",
	// ... complete the map for all squares
}

// Update the getLegalMoves function to use board.AllMoves()
func getLegalMoves() []Move {
	moves := []Move{}
	color := board.Turn

	// For each piece type
	for p := chess.Pawns; p <= chess.Kings; p++ {
		pieceBB := board.PieceBB[color][p]

		// For each piece of this type
		for pieceBB != 0 {
			// Get the location of the piece
			startSq := chess.Square(bits.TrailingZeros64(uint64(pieceBB)))
			pieceBB &= pieceBB - 1 // Clear the bit

			// Now get the valid moves for this piece
			var pieceMoves chess.Bitboard

			switch p {
			case chess.Pawns:
				pieceMoves = board.GetPawnMoves(color) | board.GetPawnAttacks(color)
			case chess.Knights:
				pieceMoves = chess.AllKnightMoves[startSq] & ^board.ColorBB[color]
			case chess.Bishops:
				pieceMoves = chess.GetBishopMoves(startSq, board.FullBB) & ^board.ColorBB[color]
			case chess.Rooks:
				pieceMoves = chess.GetRookMoves(startSq, board.FullBB) & ^board.ColorBB[color]
			case chess.Queens:
				pieceMoves = chess.GetQueenMoves(startSq, board.FullBB) & ^board.ColorBB[color]
			case chess.Kings:
				pieceMoves = chess.GetKingMoves(startSq, board.FullBB, color, board.RKRmoved[color], board.AllAttacks(color.Other())) & ^board.ColorBB[color]
			}

			// For each possible destination
			for pieceMoves != 0 {
				endSq := chess.Square(bits.TrailingZeros64(uint64(pieceMoves)))
				pieceMoves &= pieceMoves - 1 // Clear the bit

				// Check if this move is legal (considering checks, etc.)
				actualPiece, _ := board.IsLegal(startSq, endSq, chess.Empty)
				if actualPiece != chess.Empty {
					move := Move{
						Start: IndexToNotation[startSq],
						End:   IndexToNotation[endSq],
					}

					// Handle promotion for pawns
					if p == chess.Pawns && (endSq >= 56 || endSq <= 7) { // Promotion rank
						// For simplicity, always promote to queen
						// In a full implementation, you'd generate all possible promotions
						var pieces = []string{"bishop", "knight", "rook", "queen"}
						for i := range 4 {
							move.Promotion = pieces[i]
							moves = append(moves, move)
						}
						break
					}

					moves = append(moves, move)
				}
			}
		}
	}

	return moves
}

func isInCheck() bool {
	// This is a placeholder - implement check detection logic
	var otherColor = board.Turn.Other()
	if board.AllAttacks(otherColor)&board.PieceBB[board.Turn][chess.Kings] != 0 {
		return true
	}
	return false
}

func main() {
	// Initialize the board
	board = chess.NewBoard()

	// Set up the HTTP routes
	http.HandleFunc("/board", boardHandler)
	http.HandleFunc("/move", moveHandler)
	http.HandleFunc("/reset", resetHandler)

	// Start the server
	port := 8080
	fmt.Printf("Chess engine server running on port %d\n", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
