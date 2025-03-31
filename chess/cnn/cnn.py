import tensorflow as tf
from go_to_py import ChessGame

def create_chess_model():
    # Input: 8x8 board with multiple channels for piece positions
    inputs = tf.keras.Input(
        shape=(8, 8, 19)
    )  # 19 channels: 6 piece types x 2 colors + 7 additional features

    # Residual tower
    x = tf.keras.layers.Conv2D(256, 3, padding="same")(inputs)
    x = tf.keras.layers.BatchNormalization()(x)
    x = tf.keras.layers.ReLU()(x)

    # Several residual blocks
    for _ in range(19):
        residual = x
        x = tf.keras.layers.Conv2D(256, 3, padding="same")(x)
        x = tf.keras.layers.BatchNormalization()(x)
        x = tf.keras.layers.ReLU()(x)
        x = tf.keras.layers.Conv2D(256, 3, padding="same")(x)
        x = tf.keras.layers.BatchNormalization()(x)
        x = x + residual  # Skip connection
        x = tf.keras.layers.ReLU()(x)

    # Policy head (move probabilities)
    policy_head = tf.keras.layers.Conv2D(2, 1, padding="same")(x)
    policy_head = tf.keras.layers.BatchNormalization()(policy_head)
    policy_head = tf.keras.layers.ReLU()(policy_head)
    policy_head = tf.keras.layers.Flatten()(policy_head)
    policy_head = tf.keras.layers.Dense(1968, activation="softmax")(
        policy_head
    )  # 1968 possible moves

    # Value head (position evaluation)
    value_head = tf.keras.layers.Conv2D(1, 1, padding="same")(x)
    value_head = tf.keras.layers.BatchNormalization()(value_head)
    value_head = tf.keras.layers.ReLU()(value_head)
    value_head = tf.keras.layers.Flatten()(value_head)
    value_head = tf.keras.layers.Dense(256, activation="relu")(value_head)
    value_head = tf.keras.layers.Dense(1, activation="tanh")(value_head)

    model = tf.keras.Model(inputs=inputs, outputs=[policy_head, value_head])
    return model


class MCTSNode:
    def __init__(self, state, prior=0, parent=None):
        self.state = state
        self.prior = prior
        self.parent = parent
        self.children = {}
        self.visit_count = 0
        self.value_sum = 0
        self.is_expanded = False

    def select_child(self, c_puct=1.0):
        best_score = -float("inf")
        best_action = None

        # Sum of all child visit counts
        sum_visit_count = sum(child.visit_count for child in self.children.values())

        for action, child in self.children.items():
            ucb_score = child.get_value() + c_puct * child.prior * (
                sum_visit_count**0.5
            ) / (1 + child.visit_count)

            if ucb_score > best_score:
                best_score = ucb_score
                best_action = action

        return best_action, self.children[best_action]

    def expand(self, policy):
        # Called when we reach a new state - creates children based on policy network output
        self.is_expanded = True
        for action, prob in enumerate(policy):
            if prob > 0:  # Only consider legal moves with non-zero probability
                next_state = self.state.make_move(action)
                if next_state is not None:  # Valid move
                    self.children[action] = MCTSNode(
                        next_state, prior=prob, parent=self
                    )

    def get_value(self):
        if self.visit_count == 0:
            return 0
        return self.value_sum / self.visit_count


def self_play(model, num_games=100):
    training_data = []

    for _ in range(num_games):
        game = ChessGame()  # Initialize Go chess game
        states, policies, values = [], [], []

        while not game.is_terminal():
            # Run MCTS for this position
            root = MCTSNode(game)
            for _ in range(800):
                mcts_search(root, model)

            # Get policy from the visit counts
            policy = [0] * 1968  # All possible moves
            for action, child in root.children.items():
                policy[action] = child.visit_count

            # Normalize the policy
            sum_visits = sum(policy)
            policy = [count / sum_visits for count in policy]

            # Store the current state
            states.append(game.get_canonical_state())
            policies.append(policy)

            # Select move based on visit counts
            if training:
                # Use temperature parameter to control exploration
                temperature = 1.0 if game.move_count < 30 else 0.1
                action = select_action_with_temperature(policy, temperature)
            else:
                # During actual play, choose the most visited move
                action = policy.index(max(policy))

            game.make_move(action)

            # If game finished, add results
            if game.is_terminal():
                result = game.get_result()
                values = [
                    result * ((-1) ** (len(states) - 1 - i)) for i in range(len(states))
                ]

        # Add all positions from this game to training data
        for state, policy, value in zip(states, policies, values):
            training_data.append((state, policy, value))

    return training_data


def train_model(model, training_data):
    states, policies, values = zip(*training_data)
    model.fit(
        x=np.array(states),
        y=[np.array(policies), np.array(values)],
        batch_size=2048,
        epochs=10,
    )
    return model
