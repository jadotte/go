package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
)

func main() {
	var (
		name       string
		difficulty string
		pieces     []string
		confirm    bool
	)

	// Available chess pieces
	pieceOptions := []huh.Option[string]{
		{Key: "pawn", Value: "Pawn"},
		{Key: "knight", Value: "Knight"},
		{Key: "bishop", Value: "Bishop"},
		{Key: "rook", Value: "Rook"},
		{Key: "queen", Value: "Queen"},
		{Key: "king", Value: "King"},
	}

	// Difficulty levels
	difficultyOptions := []huh.Option[string]{
		{Key: "easy", Value: "Easy"},
		{Key: "medium", Value: "Medium"},
		{Key: "hard", Value: "Hard"},
	}

	// Create a form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What's your name?").
				Placeholder("Enter your name").
				Value(&name),

			huh.NewSelect[string]().
				Title("Select difficulty").
				Options(difficultyOptions...).
				Value(&difficulty),

			huh.NewMultiSelect[string]().
				Title("Select your favorite chess pieces").
				Options(pieceOptions...).
				Value(&pieces),

			huh.NewConfirm().
				Title("Are you ready to play?").
				Value(&confirm),
		),
	).WithTheme(huh.ThemeCatppuccin())

	// Run the form
	err := form.Run()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Print the results
	fmt.Println("🎮 Game Configuration Summary 🎮")
	fmt.Println("Player Name:", name)
	fmt.Println("Difficulty:", difficulty)
	fmt.Println("Favorite Pieces:", strings.Join(pieces, ", "))

	if confirm {
		fmt.Println("Game is starting! Good luck!")
	} else {
		fmt.Println("Setup complete, but you chose not to start the game.")
	}
}
