package erisplugins

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/olympus-go/eris/utils"
	"github.com/rs/zerolog/log"
)

type PingPlugin struct {
}

func Ping() PingPlugin {
	return PingPlugin{}
}

func (p PingPlugin) Name() string {
	return "Ping"
}

func (p PingPlugin) Description() string {
	return "Is it working yet?"
}

func (p PingPlugin) Handlers() map[string]any {
	handlers := make(map[string]any)

	handlers["ping_handler"] = func(session *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			applicationCommandData := i.ApplicationCommandData()

			if applicationCommandData.Name != "ping" {
				return
			}

			message := ""

			if len(applicationCommandData.Options) > 0 {
				recipientID := applicationCommandData.Options[0].Value.(string)
				recipientMessage := fmt.Sprintf("Pong! (You were pinged by <@%s>!)", utils.GetInteractionUserId(i.Interaction))

				recipientUserChannel, err := session.UserChannelCreate(recipientID)
				if err != nil {
					log.Error().Err(err).Str("userId", recipientID).Msg("could not create DM with user")
					return
				}
				_, err = session.ChannelMessageSend(recipientUserChannel.ID, recipientMessage)
				if err != nil {
					log.Error().Err(err).Str("userId", recipientID).Msg("could not send DM with user")
					return
				}

				message = fmt.Sprintf("You pinged <@%s>!", recipientID)
			} else {
				message = "Pong!"
			}

			_ = utils.SendEphemeralInteractionResponse(session, i.Interaction, message)
		}
	}
	return handlers
}

func (p PingPlugin) Commands() map[string]*discordgo.ApplicationCommand {
	commands := make(map[string]*discordgo.ApplicationCommand)

	commands["ping_command"] = &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Pong!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "user",
				Description: "Pings the specified user",
				Type:        discordgo.ApplicationCommandOptionUser,
				Required:    false,
			},
		},
	}
	return commands
}

func (p PingPlugin) Intents() []discordgo.Intent {
	return nil
}
