package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/olympus-go/eris/utils"
	"github.com/rs/zerolog/log"
	"strings"
)

type TicTacToePlugin struct {
	activeGames map[string]*ticTacToeGame
}

func TicTacToe() TicTacToePlugin {
	return TicTacToePlugin{activeGames: make(map[string]*ticTacToeGame)}
}

func (t TicTacToePlugin) Name() string {
	return "Tic-Tac-Toe"
}

func (t TicTacToePlugin) Description() string {
	return "Play tic-tac-toe with your friends!"
}

func (t TicTacToePlugin) Handlers() map[string]any {
	handlers := make(map[string]any)

	handlers["tictactoe_handler"] = func(session *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			applicationCommandData := i.ApplicationCommandData()

			if applicationCommandData.Name != "tictactoe" {
				return
			}

			message := ""

			//no need to check for options length as user is required

			if len(applicationCommandData.Options) > 0 {
				newGame := ticTacToeGame{
					challengerID: utils.GetInteractionUserId(i.Interaction),
					challengeeID: applicationCommandData.Options[0].Value.(string),
					accepted:     false,
				}
				newGame.gameID = fmt.Sprintf("%sand%s", newGame.challengerID, newGame.challengeeID)

				recipientMessage := fmt.Sprintf("You have been challenged to a Tic-Tac-Toe game by by <@%s>!", newGame.challengerID)
				recipientUserChannel, err := session.UserChannelCreate(newGame.challengeeID)
				if err != nil {
					return
				}

				messageSend := discordgo.MessageSend{
					Content: recipientMessage,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Accept",
									Style:    discordgo.PrimaryButton,
									CustomID: "ttt_challenge_accept_" + newGame.gameID,
									Disabled: false,
								},
								discordgo.Button{
									Label:    "Decline",
									Style:    discordgo.PrimaryButton,
									CustomID: "ttt_challenge_decline_" + newGame.gameID,
									Disabled: false,
								},
							},
						},
					},
				}

				_, err = session.ChannelMessageSendComplex(recipientUserChannel.ID, &messageSend)
				if err != nil {
					return
				}

				t.activeGames[newGame.gameID] = &newGame

				message = fmt.Sprintf("You challenged <@%s> to a game of tic-tac-toe!", newGame.challengeeID)
			} else {
				message = "idk how you got here"
			}

			_ = utils.SendEphemeralInteractionResponse(session, i.Interaction, message)
		case discordgo.InteractionMessageComponent:
			log.Debug().Msg(i.MessageComponentData().CustomID)
			if strings.HasPrefix(i.MessageComponentData().CustomID, "ttt_challenge_decline_") {
				gameID := strings.TrimPrefix(i.MessageComponentData().CustomID, "ttt_challenge_decline_")

				log.Debug().Msg(gameID)

				if _, ok := t.activeGames[gameID]; !ok {
					utils.InteractionResponse(session, i.Interaction).Flags(discordgo.MessageFlagsEphemeral).Message("Challenge doesn't exist.").SendWithLog(log.Logger)
					return
				} else {
					delete(t.activeGames, gameID)
					utils.InteractionResponse(session, i.Interaction).Flags(discordgo.MessageFlagsEphemeral).Message("Declined game challenge.").SendWithLog(log.Logger)
				}

			} else if strings.HasPrefix(i.MessageComponentData().CustomID, "ttt_challenge_accept_") {
				gameID := strings.TrimPrefix(i.MessageComponentData().CustomID, "ttt_challenge_accept_")
				t.activeGames[gameID].accepted = true
				log.Debug().Msg(t.activeGames[gameID].gameID)

				utils.InteractionResponse(session, i.Interaction).Message("Accepted challenge!")
			}

		}
	}
	return handlers
}

func (t TicTacToePlugin) Commands() map[string]*discordgo.ApplicationCommand {
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

func (t TicTacToePlugin) Intents() []discordgo.Intent {
	return nil
}

func sendGameRequest() {

}

type ticTacToeGame struct {
	gameID       string
	challengerID string
	challengeeID string
	accepted     bool
}
