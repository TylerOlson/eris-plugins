package erisplugins

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/olympus-go/eris/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

type TicTacToePlugin struct {
	activeGames map[string]*ticTacToeGame
	logger      zerolog.Logger
}

type ticTacToeGame struct {
	ID   string
	Data [3][3]string
	Turn string

	message   *utils.InteractionResponseBuilder
	playerIDs [2]string // x, o
	xScores   [8]int    //[row1, row2, row3, col1, col2, col3, diag1, diag2]
	oScores   [8]int    // grid * 2 + 2
}

func TicTacToe(logger zerolog.Logger) *TicTacToePlugin {
	plugin := TicTacToePlugin{
		activeGames: make(map[string]*ticTacToeGame),
		logger:      logger.With().Str("plugin", "tictactoe").Logger(),
	}
	return &plugin
}

func (t *TicTacToePlugin) Name() string {
	return "Tic-Tac-Toe"
}

func (t *TicTacToePlugin) Description() string {
	return "Play tic-tac-toe with your friends!"
}

func (t *TicTacToePlugin) Handlers() map[string]any {
	handlers := make(map[string]any)

	handlers["tictactoe_handler"] = func(session *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if i.ApplicationCommandData().Name == "tictactoe" {
				// make sure command originates from guild
				if i.Interaction.GuildID == "" {
					utils.InteractionResponse(session, i.Interaction).Message("Please use in a channel, not a DM!").Ephemeral().SendWithLog(t.logger)
					t.logger.Debug().Str("user_id", i.Interaction.User.ID).Msg(" Received dm from user")
					return
				}

				challengerID := utils.GetInteractionUserId(i.Interaction)
				opponentID := i.ApplicationCommandData().Options[0].Value.(string)
				gameID := fmt.Sprintf("%sv%s", challengerID, opponentID)

				if _, ok := t.activeGames[gameID]; ok { // don't make another game
					utils.InteractionResponse(session, i.Interaction).Message("Game already running!").Ephemeral().SendWithLog(t.logger)
					return
				}

				// create a game and display it
				t.createAndDisplayTicTacToeGame(session, i, gameID, challengerID, opponentID)
			}

		case discordgo.InteractionMessageComponent:
			// if move event
			if strings.HasPrefix(i.MessageComponentData().CustomID, "ttt_move_") {
				splitID := strings.Split(i.MessageComponentData().CustomID, "_") // ttt_move_row_colum_id
				gameID := splitID[4]

				log.Debug().Str("Received message", i.MessageComponentData().CustomID)

				if !strings.Contains(gameID, utils.GetInteractionUserId(i.Interaction)) { // make sure this is users game
					utils.InteractionResponse(session, i.Interaction).Message("Not your game!").Ephemeral().SendWithLog(t.logger)
					return
				}

				if _, ok := t.activeGames[gameID]; !ok { // make sure game is active, no seg faults for me
					utils.InteractionResponse(session, i.Interaction).Message("Game is over!").Ephemeral().SendWithLog(t.logger)
					return
				}

				// convert strings to ints
				row, err := strconv.Atoi(splitID[2])
				if err != nil {
					t.logger.Err(err).Msg("Couldn't convert row")
					return
				}
				column, err := strconv.Atoi(splitID[3])
				if err != nil {
					t.logger.Err(err).Msg("Couldn't convert column")
					return
				}

				//update data and edit the existing board message
				t.updateTicTacToeGame(gameID, row, column)

				// if there is a winner display winner message, otherwise respond with ok response
				if t.isWinner(gameID) != "" {
					t.setBoardWinner(gameID)
					utils.InteractionResponse(session, i.Interaction).Message(t.isWinner(gameID) + " wins!").SendWithLog(t.logger)
					delete(t.activeGames, gameID)
				} else {
					utils.InteractionResponse(session, i.Interaction).Type(discordgo.InteractionResponseDeferredMessageUpdate).SendWithLog(t.logger)
				}
			}
		}
	}
	return handlers
}

func (t *TicTacToePlugin) Commands() map[string]*discordgo.ApplicationCommand {
	commands := make(map[string]*discordgo.ApplicationCommand)

	commands["tictactoe_command"] = &discordgo.ApplicationCommand{
		Name:        "tictactoe",
		Description: "Challenge your friend to a good ol' game of tic-tac-toe!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "user",
				Description: "Challenges the specified user",
				Type:        discordgo.ApplicationCommandOptionUser,
				Required:    true,
			},
		},
	}
	return commands
}

func (t *TicTacToePlugin) Intents() []discordgo.Intent {
	return nil
}

func (t *TicTacToePlugin) createAndDisplayTicTacToeGame(session *discordgo.Session, i *discordgo.InteractionCreate, gameID string, challengerID string, opponentID string) {
	t.activeGames[gameID] = &ticTacToeGame{
		ID: gameID,
		Data: [3][3]string{
			{"1", "2", "3"},
			{"4", "5", "6"},
			{"7", "8", "9"},
		},
		Turn: "X",
	}

	components := t.getTicTacToeComponents(gameID, true)
	message := utils.InteractionResponse(session, i.Interaction).Components(components[0], components[1], components[2]).Message(fmt.Sprintf("<@%s> vs <@%s>", challengerID, opponentID))

	err := message.Send()
	if err != nil {
		t.logger.Err(err).Msg("Couldn't send board message")
		return
	}

	t.activeGames[gameID].message = message

	t.logger.Debug().Interface("game", t.activeGames[gameID]).Msg("created game")
}

func (t *TicTacToePlugin) updateTicTacToeGame(gameID string, row int, column int) {
	t.activeGames[gameID].Data[row][column] = t.activeGames[gameID].Turn

	if t.activeGames[gameID].Turn == "X" {
		t.activeGames[gameID].xScores[row] += 1
		t.activeGames[gameID].xScores[column+3] += 1
		if row == column {
			t.activeGames[gameID].xScores[6] += 1
		}
		if 3-1-column == row {
			t.activeGames[gameID].xScores[7] += 1
		}
		t.activeGames[gameID].Turn = "O"
	} else if t.activeGames[gameID].Turn == "O" {
		t.activeGames[gameID].Turn = "X"
		t.activeGames[gameID].oScores[row] += 1
		t.activeGames[gameID].oScores[column+3] += 1
		if row == column {
			t.activeGames[gameID].oScores[6] += 1
		}
		if 3-1-column == row {
			t.activeGames[gameID].oScores[7] += 1
		}
	}

	components := t.getTicTacToeComponents(gameID, true)

	t.activeGames[gameID].message = t.activeGames[gameID].message.Components(components[0], components[1], components[2])
	t.activeGames[gameID].message.EditWithLog(t.logger)

	t.logger.Debug().Interface("game", t.activeGames[gameID]).Msg("updated game")
}

func (t *TicTacToePlugin) setBoardWinner(gameID string) {
	components := t.getTicTacToeComponents(gameID, false)

	t.activeGames[gameID].message = t.activeGames[gameID].message.Components(components[0], components[1], components[2])
	t.activeGames[gameID].message.EditWithLog(t.logger)
}

func (t *TicTacToePlugin) getTicTacToeComponents(gameID string, enabled bool) []discordgo.MessageComponent {
	game := t.activeGames[gameID]
	var components []discordgo.MessageComponent
	for row := 0; row < 3; row++ {
		var actionsRow utils.ActionsRowBuilder
		for col := 0; col < 3; col++ {
			button := utils.Button().Label(game.Data[row][col]).Enabled(enabled).
				Id(fmt.Sprintf("ttt_move_%d_%d_%s", row, col, gameID)).Build()
			actionsRow.Button(button)
		}
		components = append(components, actionsRow.Build())
	}

	return components
}

func (t *TicTacToePlugin) isWinner(gameID string) string {
	game := t.activeGames[gameID]

	for i := 0; i < 8; i++ {
		if game.xScores[i] == 3 {
			return "X"
		} else if game.oScores[i] == 3 {
			return "O"
		}
	}

	return ""
}
