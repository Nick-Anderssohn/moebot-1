package commands

import (
	"strings"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PermitCommand struct {
}

func (pc *PermitCommand) Execute(pack *CommPackage) {
	args := ParseCommand(strings.Join(pack.params, " "), []string{"-permission", "-securityAnswer", "-confirmationMessage"})

	if _, ok := args[""]; !ok || len(args) <= 1 {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a role name followed by one or more arguments.")
		return
	}

	role := db.Role{}

	if perm, ok := args["-permission"]; ok {
		permLevel := db.GetPermissionFromString(perm)
		if permLevel == -1 {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Invalid permission level")
			return
		}
		role.Permission = permLevel
	}

	if sa, ok := args["-securityAnswer"]; ok {
		if sa == "" {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a valid security answer")
			return
		}
		role.ConfirmationSecurityAnswer.Scan(sa)
	}

	if message, ok := args["-confirmationMessage"]; ok {
		if message == "" {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a valid confirmation message")
			return
		}
		role.ConfirmationMessage.Scan(message)
	}

	// find the correct role
	roleName := args[""]
	r := util.FindRoleByName(pack.guild.Roles, roleName)
	if r == nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Unknown role name")
	}
	// we've got the role, add it to the db, updating if necessary
	// but first grab the server (probably want to move this out to include in the commPackage
	s, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Error retrieving server information")
		return
	}
	role.ServerId = s.Id
	role.RoleUid = r.ID
	db.RoleInsertOrUpdate(role)
	pack.session.ChannelMessageSend(pack.channel.ID, "Edited role "+roleName+" successfully")
}

func (pc *PermitCommand) GetPermLevel() db.Permission {
	return db.PermMaster
}

func (pc *PermitCommand) GetCommandKeys() []string {
	return []string{"PERMIT"}
}
