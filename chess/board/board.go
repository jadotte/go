package chess

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"strings"
)

type (
	Bitboard uint64 // Using bitboard to improve model efficiency
	Piece    uint8
	Color    uint8
	Square   uint8
)

const (
	Empty Piece = iota
	Pawns
	Knights
	Bishops
	Rooks
	Queens
	Kings
)

const (
	White Color = iota
	Black
	N = 16
)

var (
	allKnightMoves  = GenAllKnightMoves()
	NotationToIndex = map[string]Square{
		"a1": 0, "b1": 1, "c1": 2, "d1": 3, "e1": 4, "f1": 5, "g1": 6, "h1": 7,
		"a2": 8, "b2": 9, "c2": 10, "d2": 11, "e2": 12, "f2": 13, "g2": 14, "h2": 15,
		"a3": 16, "b3": 17, "c3": 18, "d3": 19, "e3": 20, "f3": 21, "g3": 22, "h3": 23,
		"a4": 24, "b4": 25, "c4": 26, "d4": 27, "e4": 28, "f4": 29, "g4": 30, "h4": 31,
		"a5": 32, "b5": 33, "c5": 34, "d5": 35, "e5": 36, "f5": 37, "g5": 38, "h5": 39,
		"a6": 40, "b6": 41, "c6": 42, "d6": 43, "e6": 44, "f6": 45, "g6": 46, "h6": 47,
		"a7": 48, "b7": 49, "c7": 50, "d7": 51, "e7": 52, "f7": 53, "g7": 54, "h7": 55,
		"a8": 56, "b8": 57, "c8": 58, "d8": 59, "e8": 60, "f8": 61, "g8": 62, "h8": 63,
	}
)
var AllKnightMoves = allKnightMoves
var (
	FileA Bitboard = 0x0101010101010101
	FileB Bitboard = FileA << 1
	FileC Bitboard = FileA << 2
	FileD Bitboard = FileA << 3
	FileE Bitboard = FileA << 4
	FileF Bitboard = FileA << 5
	FileG Bitboard = FileA << 6
	FileH Bitboard = FileA << 7

	Rank1 Bitboard = 0x00000000000000FF
	Rank2 Bitboard = Rank1 << 8
	Rank3 Bitboard = Rank1 << 16
	Rank4 Bitboard = Rank1 << 24
	Rank5 Bitboard = Rank1 << 32
	Rank6 Bitboard = Rank1 << 40
	Rank7 Bitboard = Rank1 << 48
	Rank8 Bitboard = Rank1 << 56
)

type Board struct {
	PieceBB [2][7]Bitboard

	ColorBB [2]Bitboard

	FullBB Bitboard

	Turn Color

	KnightMoves [64]Bitboard

	RKRmoved [2][3]bool

	EnPassantSquare *Square

	MoveCounter uint8
}

func NewBoard() *Board {
	// initializes new board with starting chess possition
	// filled in all sub-bitboards 
	b := &Board{
		Turn: White,
	}
	b.PieceBB[White][Pawns] = Rank2
	b.PieceBB[White][Knights] = (1 << NotationToIndex["b1"]) | (1 << NotationToIndex["g1"])
	b.PieceBB[White][Bishops] = (1 << NotationToIndex["c1"]) | (1 << NotationToIndex["f1"])
	b.PieceBB[White][Rooks] = (1 << NotationToIndex["a1"]) | (1 << NotationToIndex["h1"])
	b.PieceBB[White][Queens] = (1 << NotationToIndex["d1"])
	b.PieceBB[White][Kings] = (1 << NotationToIndex["e1"])

	b.PieceBB[Black][Pawns] = Rank7
	b.PieceBB[Black][Knights] = (1 << NotationToIndex["b8"]) | (1 << NotationToIndex["g8"])
	b.PieceBB[Black][Bishops] = (1 << NotationToIndex["c8"]) | (1 << NotationToIndex["f8"])
	b.PieceBB[Black][Rooks] = (1 << NotationToIndex["a8"]) | (1 << NotationToIndex["h8"])
	b.PieceBB[Black][Queens] = (1 << NotationToIndex["d8"])
	b.PieceBB[Black][Kings] = (1 << NotationToIndex["e8"])
	b.CombineBB()
	b.EnPassantSquare = nil

	return b
}

func (b *Board) IsLegal(start Square, end Square, promotion Piece) (Piece, Piece) {
	// Checks if moving piece from start to end with promotion
	// (or not if empty) is legal.
	color := b.Turn
	otherColor := color.Other()
	opponentMoves := b.AllAttacks(otherColor)
	piece := b.GetPieceAt(start, color)
	output := Empty
	outputPromotion := Empty
	var endBB Bitboard = end.GetFile() & end.GetRank()
	// checks through the pieceBB for the piece we are moving
	if promotion == Pawns || promotion == Kings {
		return Empty, Empty
	}
	switch piece {
	case Empty:
		fmt.Println("must select piece")
		return Empty, Empty
	case Pawns:
		pawnMoves := GetPawnMoves(start, b.FullBB, color, b.ColorBB[otherColor])
		if (pawnMoves & endBB & ^b.ColorBB[color]) != 0 {
			if end.GetRank() == Rank8 || end.GetRank() == Rank1 {
				if promotion != Empty {
					outputPromotion = promotion
				} else {
					return Empty, Empty
				}
			}
			output = Pawns
		}
	case Knights:
		if (allKnightMoves[start] & endBB & ^b.ColorBB[color]) != 0 {
			output = Knights
		}
	case Bishops:
		if (GetBishopMoves(start, b.FullBB) & endBB & ^b.ColorBB[color]) != 0 {
			output = Bishops
		}
	case Rooks:
		if (GetRookMoves(start, b.FullBB) & endBB & ^b.ColorBB[color]) != 0 {
			output = Rooks
		}
	case Queens:
		if (GetQueenMoves(start, b.FullBB) & endBB & ^b.ColorBB[color]) != 0 {
			output = Queens
		}
	case Kings:
		if (GetKingMoves(start, b.FullBB, color, b.RKRmoved[color], opponentMoves) & endBB & ^b.ColorBB[color] & ^opponentMoves) != 0 {
			output = Kings
		}
	}
	// king can't be in check after move
	if (b.PieceBB[color][Kings] & opponentMoves) != 0 {
		fmt.Println("You are in check")
		fmt.Println(PrintBB(opponentMoves))
		return Empty, Empty
	} else if promotion != outputPromotion {
		return Empty, Empty
	}
	return output, outputPromotion
}

func (b *Board) MovePiece(piece Piece, start, end Square, promotion Piece) bool {
	// moves piece at start to end and promotes if specified
	// returns true if there is a piece captured
	color := b.Turn
	otherColor := color.Other()
	startFile := start.GetFile()
	startRank := start.GetRank()
	endFile := end.GetFile()
	endRank := end.GetRank()

	var capture bool = false

	if piece == Empty {
		return false
	}
	// checks through the other colored bb to remove captured piece if relevent
	for p := Pawns; p <= Kings; p++ {
		if b.PieceBB[otherColor][p]&(1<<end) != 0 {
			b.PieceBB[otherColor][p].ZeroBit(end)
			capture = true
		}
	}
	// checks if castling
	if piece == Kings {
		if startFile == FileE && endFile == FileC {
			b.MoveCounter--
			if color == White {
				b.MovePiece(Rooks, 1, 4, Empty)
			} else {
				b.MovePiece(Rooks, 56, 59, Empty)
			}
			b.RKRmoved[color][2] = true
		} else if startFile == FileE && endFile == FileG {
			b.MoveCounter--
			if color == White {
				b.MovePiece(Rooks, 7, 5, Empty)
			} else {
				b.MovePiece(Rooks, 63, 61, Empty)
			}
			b.RKRmoved[color][0] = true
		}
		b.RKRmoved[color][1] = true
	} else if piece == Rooks {
		if startFile == FileA {
			b.RKRmoved[color][0] = true
		} else if startFile == FileH {
			b.RKRmoved[color][2] = true
		}
	} else if piece == Pawns {
		if color == White && startRank == Rank2 && endRank == Rank4 {
			newsq := end - 8
			b.EnPassantSquare = &newsq
		} else if color == Black && startRank == Rank7 && endRank == Rank5 {
			newsq := end + 8
			b.EnPassantSquare = &newsq
		} else {
			b.EnPassantSquare = nil
		}
	}
	// actually moves the selected piece with promotion check
	if promotion == Empty {
		b.PieceBB[color][piece].SetBit(end)
	} else {
		b.PieceBB[color][promotion].SetBit(end)
	}
	b.PieceBB[color][piece].ZeroBit(start)
	// pushes through the changed piecebb to affect all other bbs
	b.CombineBB()
	b.MoveCounter++
	return capture
}

func (b *Board) CombineBB() {
	b.ColorBB[White] = 0
	b.ColorBB[Black] = 0

	for p := Pawns; p <= Kings; p++ {
		b.ColorBB[White] |= b.PieceBB[White][p]
		b.ColorBB[Black] |= b.PieceBB[Black][p]
	}

	b.FullBB = b.ColorBB[White] | b.ColorBB[Black]
}

func (bb Bitboard) GetBit(sq Square) bool {
	return (bb & (1 << sq)) != 0
}

func (bb *Bitboard) ZeroBit(sq Square) {
	*bb &= ^(1 << sq)
}

func (bb *Bitboard) SetBit(sq Square) {
	*bb |= 1 << sq
}

func (c Color) Other() Color {
	if c == White {
		return Black
	} else {
		return White
	}
}

func (b *Board) GetPieceAt(sq Square, color Color) Piece {
	for p := Pawns; p <= Kings; p++ {
		if b.PieceBB[color][p].GetBit(sq) {
			return p
		}
	}
	return Empty
}

func (b *Board) AllMoves(color Color) Bitboard {
	var moves = b.GetPawnAttacks(color)
	moves |= b.GetPawnMoves(color) | b.GetKnightMoves(color)
	moves |= b.GetBishopMoves(color) | b.GetRookMoves(color)
	moves |= b.GetQueenMoves(color) | b.GetKingMoves(color)
	return moves
}

func (b *Board) AllAttacks(color Color) Bitboard {
	var moves Bitboard
	moves |= b.GetPawnAttacks(color) | b.GetKnightMoves(color)
	moves |= b.GetBishopMoves(color) | b.GetRookMoves(color)
	moves |= b.GetQueenMoves(color) | b.GetKingMoves(color)
	return moves
}

func (sq Square) GetRank() Bitboard {
	ranks := []Bitboard{Rank1, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8}
	return ranks[sq/8]
}

func (sq Square) GetFile() Bitboard {
	files := []Bitboard{FileA, FileB, FileC, FileD, FileE, FileF, FileG, FileH}
	return files[sq%8]
}

func getSymbol(c Color, p Piece) rune {
	// returns coresponding symbols for piece/color combo inputted.
	// White is uppercase and black is lowercase

	var sym rune
	switch p {
	case Empty:
		sym = '.'
	case Pawns:
		sym = 'P'
	case Knights:
		sym = 'N'
	case Bishops:
		sym = 'B'
	case Rooks:
		sym = 'R'
	case Queens:
		sym = 'Q'
	case Kings:
		sym = 'K'
	}
	if c == Black && sym != '.' {
		sym += 32
	}
	return sym
}

func (b *Board) PrintBoard() string {
	var sb strings.Builder
	sb.Grow(90)
	pieces := [64]rune{}
	for i := range pieces {
		pieces[i] = '.'
	}
	for c := White; c <= Black; c++ {
		for p := Empty; p <= Kings; p++ {
			bb := b.PieceBB[c][p]
			var sym rune = getSymbol(c, p)
			for bb != 0 {
				sq := Square(bits.TrailingZeros64(uint64(bb)))
				pieces[sq] = sym
				bb &= bb - 1
			}
		}
	}
	for rank := 7; rank >= 0; rank-- {
		sb.WriteString(fmt.Sprintf("%d ", rank+1))

		for file := range 8 {
			sq := Square(rank*8 + file)
			sb.WriteString(string(pieces[sq]) + " ")
		}

		sb.WriteString("\n")
	}

	sb.WriteString("  a b c d e f g h\n")

	return sb.String()
}

func PrintBB(bb Bitboard) string {
	var sb strings.Builder
	sb.Grow(90)
	pieces := [64]rune{}
	for i := range pieces {
		pieces[i] = '.'
	}
	var sym rune = 'x'
	for bb != 0 {
		sq := Square(bits.TrailingZeros64(uint64(bb)))
		pieces[sq] = sym
		bb &= bb - 1
	}

	for rank := 7; rank >= 0; rank-- {
		sb.WriteString(fmt.Sprintf("%d ", rank+1))

		for file := range 8 {
			sq := Square(rank*8 + file)
			sb.WriteString(string(pieces[sq]) + " ")
		}

		sb.WriteString("\n")
	}

	sb.WriteString("  a b c d e f g h\n")

	return sb.String()
}

func onBoard(start int, shift int) bool {
	// Checks if a piece at the start cooardinate can move
	// by the shift value without moving off the board.
	// So the end cooardinate must be within 0 and 63 and
	// cannot try crossing one of the edges.
	// For example a Knight on a2 cannot make any moves left.

	if (start+shift >= 0) && (start+shift < 64) && (start%8+shift%8 < 8) && (start%8+shift%8 >= 0) {
		return true
	}
	return false
}

func GenAllKnightMoves() [64]Bitboard {
	// Generates an array of bitboards where each bitbaord i
	// contains all of the valid moves for a knight starting at
	// cooardinate i if they were on an empty board.

	var bbs [64]Bitboard
	var dists [8]int = [8]int{-17, -15, -10, -6, 6, 10, 15, 17}
	for sq := Square(0); sq < 64; sq++ {
		rank := sq.GetRank()
		file := sq.GetFile()
		for _, dist := range dists {
			if (rank == Rank1 && dist <= -6) || (rank == Rank2 && dist <= -15) {
				continue
			} else if rank == Rank8 && dist >= 6 || (rank == Rank7 && dist >= 15) {
				continue
			} else if (file == FileA || file == FileB) && (dist == 6 || dist == -10) {
				continue
			} else if (file == FileA) && (dist == 15 || dist == -17) {
				continue
			} else if (file == FileH || file == FileG) && (dist == -6 || dist == 10) {
				continue
			} else if (file == FileH) && (dist == 17 || dist == -15) {
				continue
			}

			end := int(sq) + dist
			bbs[sq] |= 1 << uint(end)
		}
	}
	return bbs
}

func GetPawnMoves(sq Square, fullBB Bitboard, color Color, otherColorBB Bitboard) Bitboard {
	var moves Bitboard
	if color == White {
		// adds single push
		moves = (1 << (sq + 8)) & ^fullBB
		// adds double push
		moves |= ((moves & Rank3) << 8) & ^fullBB
		// adds takes
		moves |= (((1<<sq + 7) & ^FileA) & otherColorBB) | (((1<<sq + 9) & ^FileH) & otherColorBB)
	} else {
		// same for black
		moves = (1 << (sq - 8)) & ^fullBB
		moves |= ((moves & Rank6) >> 8) & ^fullBB
		moves |= (((1 << (sq - 7)) & ^FileA) & otherColorBB) | (((1 << (sq - 9)) & ^FileH) & otherColorBB)
	}
	return moves
}

func (b *Board) GetPawnMoves(color Color) Bitboard {
	// creates a bitboard of every legal move that a pawn of the
	// coresponding color could make (excluding attacks).
	var moves Bitboard
	pawns := b.PieceBB[color][Pawns]

	if color == White {
		// adds single push
		moves = (pawns << 8) & ^b.FullBB
		// adds double push
		moves |= ((moves & Rank3) << 8) & ^b.FullBB
	} else {
		// same for black
		moves = (pawns >> 8) & ^b.FullBB
		moves |= ((moves & Rank6) >> 8) & ^b.FullBB
	}
	return moves
}

func (b *Board) GetPawnAttacks(color Color) Bitboard {
	// creates a bitboard of every legal attack that a pawn of the
	// coresponding color could make.
	var moves Bitboard
	pawns := b.PieceBB[color][Pawns]
	if color == White {
		moves = (((pawns << 7) & ^FileA) & b.ColorBB[Black]) | (((pawns << 9) & ^FileH) & b.ColorBB[Black])
	} else {
		moves = (((pawns >> 7) & ^FileA) & b.ColorBB[White]) | (((pawns >> 9) & ^FileH) & b.ColorBB[White])
	}
	return moves
}

func (b *Board) GetKnightMoves(color Color) Bitboard {
	// Creates a bitboard of every legal move that a knight
	// of the coresponding color and square could make.
	knights := b.PieceBB[color][Knights]
	var moves Bitboard
	for knights != 0 {
		loc := Square(bits.TrailingZeros64(uint64(knights)))
		knights &= knights - 1
		moves |= allKnightMoves[loc] & ^b.ColorBB[color]
	}
	return moves
}

func GetBishopMoves(sq Square, fullBB Bitboard) Bitboard {
	// Creates a bitboard of every legal move that a rook
	// of the coresponding color and square could make.
	rank := sq / 8
	file := sq % 8
	var moves Bitboard = 0

	// up right
	f := file + 1
	for r := rank + 1; r < 8; r++ {
		target := r*8 + f
		moves |= 1 << uint(target)
		if (fullBB & (1 << uint(target))) != 0 {
			break
		}
		if f >= 7 {
			break
		}
		f++
	}
	// down right
	if rank != 0 {
		f = file + 1
		for r := rank - 1; r >= 0; r-- {
			target := r*8 + f
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
			if f >= 7 {
				break
			}
			f++
		}
	}
	// down left
	if rank != 0 && file != 0 {
		f = file - 1
		for r := rank - 1; r >= 0; r-- {
			target := r*8 + f
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
			if f == 0 {
				break
			}
			f--
		}
	}
	// up left
	if file != 0 {
		f = file - 1
		for r := rank + 1; r < 8; r++ {
			target := r*8 + f
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
			if f == 0 {
				break
			}
			f--
		}
	}
	return moves
}

func (b *Board) GetBishopMoves(color Color) Bitboard {
	// creates a bitboard of every legal attack that a bishop of the
	// coresponding color could make.
	bishops := b.PieceBB[color][Bishops]
	var moves Bitboard
	for bishops != 0 {
		loc := Square(bits.TrailingZeros64(uint64(bishops)))
		bishops &= bishops - 1
		moves |= GetBishopMoves(loc, b.FullBB) & ^b.ColorBB[color]
	}
	return moves
}

func GetRookMoves(sq Square, fullBB Bitboard) Bitboard {
	// Creates a bitboard of every legal move that a rook
	// of the coresponding color and square could make.
	rank := sq / 8
	file := sq % 8
	var moves Bitboard = 0

	// up
	for r := rank + 1; r < 8; r++ {
		target := r*8 + file
		moves |= 1 << uint(target)
		if (fullBB & (1 << uint(target))) != 0 {
			break
		}
	}
	// right
	for f := file + 1; f < 8; f++ {
		target := rank*8 + f
		moves |= 1 << uint(target)
		if (fullBB & (1 << uint(target))) != 0 {
			break
		}
	}
	// down
	if rank != 0 {
		for r := rank - 1; r >= 0; r-- {
			target := r*8 + file
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
		}
	}
	// left
	if file != 0 {
		for f := file - 1; f >= 0; f-- {
			target := rank*8 + f
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
		}
	}
	return moves
}

func (b *Board) GetRookMoves(color Color) Bitboard {
	rooks := b.PieceBB[color][Rooks]
	var moves Bitboard
	for rooks != 0 {
		loc := Square(bits.TrailingZeros64(uint64(rooks)))
		rooks &= rooks - 1
		moves |= GetRookMoves(loc, b.FullBB) & ^b.ColorBB[color]
	}
	return moves
}

func GetQueenMoves(sq Square, fullBB Bitboard) Bitboard {
	return GetRookMoves(sq, fullBB) | GetBishopMoves(sq, fullBB)
}

func (b *Board) GetQueenMoves(color Color) Bitboard {
	queens := b.PieceBB[color][Queens]
	var moves Bitboard
	for queens != 0 {
		loc := Square(bits.TrailingZeros64(uint64(queens)))
		queens &= queens - 1
		moves |= GetQueenMoves(loc, b.FullBB) & ^b.ColorBB[color]
	}
	return moves
}

func GetKingMoves(sq Square, fullBB Bitboard, color Color, RKR [3]bool, opBB Bitboard) Bitboard {
	king := Bitboard(1) << sq
	moves := (king << 8) |
		(king >> 8) |
		((king << 1) & ^FileA) |
		((king >> 1) & ^FileH) |
		((king << 9) & ^FileA) |
		((king << 7) & ^FileH) |
		((king >> 7) & ^FileA) |
		((king >> 9) & ^FileH)
	return moves | GetCastles(color, fullBB, RKR, opBB)
}

func GetCastles(color Color, fullBB Bitboard, RKR [3]bool, opBB Bitboard) Bitboard {
	var moves Bitboard
	var rank Bitboard
	if color == White {
		rank = Rank1
	} else {
		rank = Rank8
	}
	kingside := []Bitboard{FileF, FileG}
	queenside := []Bitboard{FileD, FileC, FileB}
	if !RKR[1] {
		if !RKR[0] {
			for i := range queenside {
				if (queenside[i]&rank&fullBB) != 0 || (queenside[i]&rank&opBB != 0) {
					break
				} else {
					moves |= FileC & rank
				}
			}
		}
		if !RKR[2] {
			for i := range kingside {
				if (kingside[i]&rank&fullBB) != 0 || (kingside[i]&rank&opBB != 0) {
					break
				} else {
					moves |= FileG & rank
				}
			}
		}
	}
	return moves
}

func (b *Board) GetKingMoves(color Color) Bitboard {
	king := b.PieceBB[color][Kings]
	var moves Bitboard
	loc := Square(bits.TrailingZeros64(uint64(king)))
	// treats opBB as empty here
	moves |= GetKingMoves(loc, b.FullBB, color, b.RKRmoved[color], 0) & ^b.ColorBB[color]
	return moves
}

func (board *Board) ToTensor() [8][8][19]float32 {
  // converts board to tensor to be used as input for CNN.
  // not yet fully implemented

	var tensor [8][8][19]float32

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			square := Square(row*8 + col)
			color := board.Turn
			piece := board.GetPieceAt(square, color)

			if color == White {
				switch piece {
				case Pawns:
					tensor[row][col][0] = 1.0
				case Knights:
					tensor[row][col][1] = 1.0
				case Bishops:
					tensor[row][col][2] = 1.0
				case Rooks:
					tensor[row][col][3] = 1.0
				case Queens:
					tensor[row][col][4] = 1.0
				case Kings:
					tensor[row][col][5] = 1.0
				}
			} else if color == Black {
				switch piece {
				case Pawns:
					tensor[row][col][6] = 1.0
				case Knights:
					tensor[row][col][7] = 1.0
				case Bishops:
					tensor[row][col][8] = 1.0
				case Rooks:
					tensor[row][col][9] = 1.0
				case Queens:
					tensor[row][col][10] = 1.0
				case Kings:
					tensor[row][col][11] = 1.0
				}
			}

			// Channel 12: Side to move (1 if white, 0 if black)
			switch color {
			case White:
				tensor[row][col][12] = 1.0
			case Black:
				tensor[row][col][12] = 0.0
			}
			// Channel 13-16: Castling rights
			opBB := board.AllAttacks(color.Other())
			whiteCastles := GetCastles(color, board.FullBB, board.RKRmoved[White], opBB)
			blackCastles := GetCastles(color, board.FullBB, board.RKRmoved[Black], opBB)
			if whiteCastles&FileG != 0 {
				tensor[row][col][13] = 1.0
			} else {
				tensor[row][col][13] = 1.0
			}
			if whiteCastles&FileC != 0 {
				tensor[row][col][14] = 1.0
			} else {
				tensor[row][col][14] = 1.0
			}
			if blackCastles&FileG != 0 {
				tensor[row][col][15] = 1.0
			} else {
				tensor[row][col][15] = 1.0
			}
			if blackCastles&FileC != 0 {
				tensor[row][col][16] = 1.0
			} else {
				tensor[row][col][16] = 1.0
			}

			// Channel 17: En passant square
			if board.EnPassantSquare != nil &&
				board.EnPassantSquare.GetRank() == Square(row).GetRank() &&
				board.EnPassantSquare.GetFile() == Square(col).GetFile() {
				tensor[row][col][17] = 1.0
			}

			// Channel 18: Move counter (normalized)
			tensor[row][col][18] = float32(board.MoveCounter) / 100.0
		}
	}

	return tensor
}

func (board *Board) ExportTensorAsJSON() []byte {
  // exports Tensor for model
	tensor := board.ToTensor()
	jsonData, err := json.Marshal(tensor)
	if err != nil {
		return nil
	}
	return jsonData
}
