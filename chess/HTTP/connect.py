import requests
import numpy as np


class GoChessConnector:
    def __init__(self, url="http://localhost:8080"):
        self.url = url

    def get_board_state(self):
        response = requests.get(f"{self.url}/board")
        data = response.json()
        return (
            np.array(data["board_tensor"]),
            data["legal_moves"],
            data["is_terminal"],
            data["game_result"],
        )

    def make_move(self, start, end, promotion=None):
        """
        Make a move on the board.
        Args:
        start (str): Starting square in algebraic notation (e.g., "e2")
        end (str): Ending square in algebraic notation (e.g., "e4")
        promotion (str, optional): Piece to promote to ("queen", "rook", "bishop", "knight")
        Returns:
            tuple: (board_repr, turn, legal_moves, is_check, board_tensor)
        """
        move_data = {"start": start, "end": end}
        if promotion:
            move_data["promotion"] = promotion

        response = requests.post(f"{self.base_url}/move", json=move_data)
        if response.status_code != 200:
            raise Exception(f"Error making move: {response.text}")

        data = response.json()
        return (
            data["board"],
            data["turn"],
            data[
                "legalMoves"
            ],  # Now a list of move objects with "start" and "end" keys
            data["isCheck"],
            np.array(data["boardTensor"]),
        )
