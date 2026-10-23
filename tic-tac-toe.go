package main

import (
	"fmt"
)

const O = " O "
const X = " x "

var currentPlayer string

var board [3][3]string

func init() {
	board = [3][3]string{{"   ", "   ", "   "}, {"   ", "   ", "   "}, {"   ", "   ", "   "}}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			fmt.Printf("[%s]", board[i][j])
		}
		fmt.Printf("\n")

	}

}

func displayBoard() {
	fmt.Printf("\n")
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			fmt.Printf("[%s]", board[i][j])
		}
		fmt.Printf("\n")

	}
}
func isBoardFull() bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] == "   " {
				return false
			}
		}

	}
	return true
}

func isValidPosition(i, j int) bool {
	if (i <= 2 && i >= 0) && (j <= 2 && j >= 0) {

		if board[i][j] == "   " {
			return true
		}
	}

	return false

}
func makeMove(i, j int, player string) error {
	if !isValidPosition(i, j) {
		return fmt.Errorf("Invalid Move")

	}
	board[i][j] = player

	return nil
}
func IsWinner(player string) bool {
	for i := 0; i < 3; i++ {
		if board[i][0] == player && board[i][1] == player && board[i][2] == player {
			return true
		}
	}
	for i := 0; i < 3; i++ {
		if board[0][i] == player && board[1][i] == player && board[2][i] == player {
			return true
		}
	}
	if board[0][0] == player && board[1][1] == player && board[2][2] == player {
		return true
	}
	if board[2][0] == player && board[1][1] == player && board[0][2] == player {
		return true
	}
	return false
}

func checkWinner() bool {
	if IsWinner(O) {
		fmt.Printf("%s win!..", O)
		return true
	} else if IsWinner(X) {
		fmt.Printf("%s Win!..", X)
		return true
	} else {
		if isBoardFull() {
			fmt.Println("Draw ...")
			return true
		}
	}
	return false
}

func main() {

	currentPlayer = O
	for !checkWinner() {

		fmt.Printf("\nEnter a move of %s player:", currentPlayer)
		var i, j int
		fmt.Scan(&i, &j)

		err := makeMove(i, j, currentPlayer)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Try again!...")
			continue
		}

		displayBoard()

		if currentPlayer == O {
			currentPlayer = X
		} else {
			currentPlayer = O
		}
	}

}
