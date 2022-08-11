package erisplugins

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/olympus-go/eris/utils"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

type TicTacToePlugin struct {
	activeGames map[string]*ticTacToeGame
}

func TicTacToe() *TicTacToePlugin {
	plugin := TicTacToePlugin{activeGames: make(map[string]*ticTacToeGame)}
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
					utils.InteractionResponse(session, i.Interaction).Message("Please use in a channel, not a DM!").SendWithLog(log.Logger)
					return
				}

				t.createTicTacToeGame(i.Interaction.ChannelID)

				t.displayTicTacToeGame(session, i, i.Interaction.ChannelID)
			}

		case discordgo.InteractionMessageComponent:
			log.Debug().Msg("Received message " + i.MessageComponentData().CustomID)

			// if move event
			if strings.HasPrefix(i.MessageComponentData().CustomID, "ttt_game_move_") {
				splitID := strings.Split(i.MessageComponentData().CustomID, "_")
				gameID := splitID[5]

				row, err := strconv.Atoi(splitID[3])
				if err != nil {
					log.Err(err)
					return
				}
				column, err := strconv.Atoi(splitID[4])
				if err != nil {
					log.Err(err)
					return
				}

				t.updateTicTacToeData(gameID, row, column, "X")
				t.updateTicTacToeBoard(gameID)
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
				Required:    false,
			},
		},
	}
	return commands
}

func (t *TicTacToePlugin) Intents() []discordgo.Intent {
	return nil
}

func (t *TicTacToePlugin) createTicTacToeGame(guildID string) {
	game := ticTacToeGame{}
	game.gameID = guildID
	game.channelID = guildID
	game.gameData = [3][3]string{
		{"1", "2", "3"},
		{"4", "5", "6"},
		{"7", "8", "9"},
	}

	t.activeGames[game.gameID] = &game
}

func (t *TicTacToePlugin) updateTicTacToeData(id string, row int, column int, move string) {
	t.activeGames[id].gameData[row][column] = move
}

func (t *TicTacToePlugin) updateTicTacToeBoard(id string) {
	components := t.getTicTacToeComponents(id)

	t.activeGames[id].message = t.activeGames[id].message.Components(components[0], components[1], components[2])
	t.activeGames[id].message.EditWithLog(log.Logger)
}

func (t *TicTacToePlugin) getTicTacToeComponents(ID string) []discordgo.MessageComponent {
	game := t.activeGames[ID]
	var components []discordgo.MessageComponent
	for row := 0; row < 3; row++ {
		var actionsRow utils.ActionsRowBuilder
		for col := 0; col < 3; col++ {
			button := utils.Button().Label(game.gameData[row][col]).
				Id(fmt.Sprintf("ttt_game_move_%d_%d_%s", row, col, game.gameID)).Build()
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
		log.Err(err).Msg("Couldn't send board message")
		return
	}

	t.activeGames[ID].message = message
}

//func sendGameRequest() {
//		newGame := ticTacToeGame{
//			challengerID: utils.GetInteractionUserId(i.Interaction),
//			challengeeID: applicationCommandData.Options[0].Value.(string),
//			accepted:     false,
//		}
//		newGame.gameID = fmt.Sprintf("%sand%s", newGame.challengerID, newGame.challengeeID)
//
//		recipientMessage := fmt.Sprintf("You have been challenged to a Tic-Tac-Toe game by <@%s>!", newGame.challengerID)
//		recipientUserChannel, err := session.UserChannelCreate(newGame.challengeeID)
//		if err != nil {
//			return
//		}
//
//		messageSend := discordgo.MessageSend{
//			Content: recipientMessage,
//			Components: []discordgo.MessageComponent{
//				discordgo.ActionsRow{
//					Components: []discordgo.MessageComponent{
//						discordgo.Button{
//							Label:    "Accept",
//							Style:    discordgo.PrimaryButton,
//							CustomID: "ttt_challenge_accept_" + newGame.gameID,
//							Disabled: false,
//						},
//						discordgo.Button{
//							Label:    "Decline",
//							Style:    discordgo.PrimaryButton,
//							CustomID: "ttt_challenge_decline_" + newGame.gameID,
//							Disabled: false,
//						},
//					},
//				},
//			},
//		}
//
//		_, err = session.ChannelMessageSendComplex(recipientUserChannel.ID, &messageSend)
//		if err != nil {
//			return
//		}
//
//		t.activeGames[newGame.gameID] = &newGame
//
//		message = fmt.Sprintf("You challenged <@%s> to a game of tic-tac-toe!", newGame.challengeeID)
//
//	_ = utils.SendEphemeralInteractionResponse(session, i.Interaction, message)
//}

type ticTacToeGame struct {
	gameID       string
	challengerID string
	gameData     [3][3]string
	channelID    string
	turn         string
	message      *utils.InteractionResponseBuilder
	//challengeeID string
	//accepted     bool
}

//log.Debug().Msg(i.MessageComponentData().CustomID)
//if strings.HasPrefix(i.MessageComponentData().CustomID, "ttt_challenge_decline_") {
//gameID := strings.TrimPrefix(i.MessageComponentData().CustomID, "ttt_challenge_decline_")
//
//log.Debug().Msg(gameID)
//
//if _, ok := t.activeGames[gameID]; !ok {
//utils.InteractionResponse(session, i.Interaction).Flags(discordgo.MessageFlagsEphemeral).Message("Challenge doesn't exist.").SendWithLog(log.Logger)
//return
//} else {
//delete(t.activeGames, gameID)
//utils.InteractionResponse(session, i.Interaction).Flags(discordgo.MessageFlagsEphemeral).Message("Declined game challenge.").SendWithLog(log.Logger)
//}
//
//} else if strings.HasPrefix(i.MessageComponentData().CustomID, "ttt_challenge_accept_") {
//gameID := strings.TrimPrefix(i.MessageComponentData().CustomID, "ttt_challenge_accept_")
//
//if _, ok := t.activeGames[gameID]; !ok {
//utils.InteractionResponse(session, i.Interaction).Flags(discordgo.MessageFlagsEphemeral).Message("Challenge doesn't exist.").SendWithLog(log.Logger)
//return
//} else {
//utils.InteractionResponse(session, i.Interaction).Flags(discordgo.MessageFlagsEphemeral).Message("Accepted challenge").SendWithLog(log.Logger)
//}
//}
