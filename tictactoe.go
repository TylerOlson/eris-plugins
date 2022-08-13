package erisplugins

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/olympus-go/eris/utils"
	"github.com/rs/zerolog"
	"strconv"
	"strings"
)

type TicTacToePlugin struct {
	activeGames map[string]*ticTacToeGame
	logger      zerolog.Logger
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

				ID := i.Interaction.ChannelID

				if _, ok := t.activeGames[ID]; ok { // don't make another game
					utils.InteractionResponse(session, i.Interaction).Message("Game already running!").Ephemeral().SendWithLog(t.logger)
					return
				}

				// create a game and display it
				t.createTicTacToeGame(ID)
				t.displayTicTacToeGame(session, i, ID)

				t.logger.Debug().Interface("game", t.activeGames[ID]).Str("user_id", ID).Msg("created game")
			}

		case discordgo.InteractionMessageComponent:
			// if move event
			if strings.HasPrefix(i.MessageComponentData().CustomID, "ttt_game_move_") {
				splitID := strings.Split(i.MessageComponentData().CustomID, "_") // ttt_game_move_row_colum_id
				gameID := splitID[5]

				if _, ok := t.activeGames[gameID]; !ok { // make sure game is active, no seg faults for me
					utils.InteractionResponse(session, i.Interaction).Message("Game is over!").Ephemeral().SendWithLog(t.logger)
					return
				}

				row, err := strconv.Atoi(splitID[3])
				if err != nil {
					t.logger.Err(err)
					return
				}
				column, err := strconv.Atoi(splitID[4])
				if err != nil {
					t.logger.Err(err)
					return
				}

				//update data and edit the existing board message, respond with ok response
				t.updateTicTacToeGame(gameID, row, column)

				if t.isWinner(gameID) != "" {
					utils.InteractionResponse(session, i.Interaction).Message(t.isWinner(gameID) + " wins!").SendWithLog(t.logger)
					delete(t.activeGames, gameID)
				} else {
					utils.InteractionResponse(session, i.Interaction).Type(discordgo.InteractionResponseDeferredMessageUpdate).SendWithLog(t.logger)
				}

				t.logger.Debug().Interface("game", t.activeGames[gameID]).Msg("updated game")
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
		//Options: []*discordgo.ApplicationCommandOption{
		//	{
		//		Name:        "user",
		//		Description: "Challenges the specified user",
		//		Type:        discordgo.ApplicationCommandOptionUser,
		//		Required:    false,
		//	},
		//},
	}
	return commands
}

func (t *TicTacToePlugin) Intents() []discordgo.Intent {
	return nil
}

func (t *TicTacToePlugin) createTicTacToeGame(guildID string) {
	t.activeGames[guildID] = &ticTacToeGame{
		ID: guildID,
		Data: [3][3]string{
			{"1", "2", "3"},
			{"4", "5", "6"},
			{"7", "8", "9"},
		},
		Turn: "X",
	}
}

func (t *TicTacToePlugin) updateTicTacToeGame(id string, row int, column int) {
	t.activeGames[id].Data[row][column] = t.activeGames[id].Turn

	if t.activeGames[id].Turn == "X" {
		t.activeGames[id].xScores[row] += 1
		t.activeGames[id].xScores[column+3] += 1
		if row == column {
			t.activeGames[id].xScores[6] += 1
		}
		if 3-1-column == row {
			t.activeGames[id].xScores[7] += 1
		}
		t.activeGames[id].Turn = "O"
	} else if t.activeGames[id].Turn == "O" {
		t.activeGames[id].Turn = "X"
		t.activeGames[id].oScores[row] += 1
		t.activeGames[id].oScores[column+3] += 1
		if row == column {
			t.activeGames[id].oScores[6] += 1
		}
		if 3-1-column == row {
			t.activeGames[id].oScores[7] += 1
		}
	}

	components := t.getTicTacToeComponents(id)

	t.activeGames[id].message = t.activeGames[id].message.Components(components[0], components[1], components[2])
	t.activeGames[id].message.EditWithLog(t.logger)

}

func (t *TicTacToePlugin) getTicTacToeComponents(ID string) []discordgo.MessageComponent {
	game := t.activeGames[ID]
	var components []discordgo.MessageComponent
	for row := 0; row < 3; row++ {
		var actionsRow utils.ActionsRowBuilder
		for col := 0; col < 3; col++ {
			button := utils.Button().Label(game.Data[row][col]).
				Id(fmt.Sprintf("ttt_game_move_%d_%d_%s", row, col, game.ID)).Build()
			actionsRow.Button(button)
		}
		components = append(components, actionsRow.Build())
	}

	return components
}

func (t *TicTacToePlugin) displayTicTacToeGame(session *discordgo.Session, i *discordgo.InteractionCreate, ID string) {
	components := t.getTicTacToeComponents(ID)

	message := utils.InteractionResponse(session, i.Interaction).Components(components[0], components[1], components[2])

	err := message.Send()
	if err != nil {
		t.logger.Err(err).Msg("Couldn't send board message")
		return
	}

	t.activeGames[ID].message = message
}

func (t *TicTacToePlugin) isWinner(ID string) string {
	game := t.activeGames[ID]

	for i := 0; i < 8; i++ {
		if game.xScores[i] == 3 {
			return "X"
		} else if game.oScores[i] == 3 {
			return "O"
		}
	}

	return ""
}

type ticTacToeGame struct {
	ID      string
	Data    [3][3]string
	Turn    string
	message *utils.InteractionResponseBuilder
	xScores [8]int //[row1, row2, row3, col1, col2, col3, diag1, diag2]
	oScores [8]int // grid * 2 + 2
}
