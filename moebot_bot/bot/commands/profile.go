package commands

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type ProfileCommand struct {
	MasterId string
}

func (pc *ProfileCommand) Execute(pack *CommPackage) {
	// special stuff for master rank
	if pc.MasterId == pack.message.Author.ID && len(pack.params) == 0 {
		pack.session.ChannelMessageSend(pack.message.ChannelID, pack.message.Author.Mention()+"'s profile:\nMy favorite user! ❤️")
		return
	}

	// technically we'll already have a user + server at this point, but may not have a usr. Still create if necessary
	server, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the server. This is an issue with moebot and not Discord.")
		return
	}
	_, err = db.UserQueryOrInsert(pack.message.Author.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching your user. This is an issue with moebot and not Discord.")
		return
	}
	usr, err := db.UserServerRankQuery(pack.message.Author.ID, pack.guild.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an issue getting your information!")
			return
		} else {
			// ErrNoRows. Overwrite the usr value, so we don't accidentally get an NPE later
			// We don't want to bail out here since this just means the user doesn't exist
			usr = nil
		}
	}
	var message strings.Builder
	message.WriteString(pack.message.Author.Mention())
	message.WriteString("'s profile:")
	message.WriteString("\nRank: ")
	if usr != nil {
		message.WriteString(convertRankToString(usr.Rank, server.VeteranRank))
	} else {
		message.WriteString("Unranked")
	}
	message.WriteString("\nPermission Level: ")
	message.WriteString(util.MakeStringCode(pc.getPermissionLevel(pack)))
	message.WriteString("\nServer join date: ")
	t, err := pack.member.JoinedAt.Parse()
	if err != nil {
		message.WriteString(util.MakeStringCode("Unknown"))
		log.Println("Problem converting server join date to time. User ID {"+pack.message.Author.ID+"}, Joined at time: {"+
			string(pack.member.JoinedAt)+"} error: ", err)
	} else {
		message.WriteString(util.MakeStringCode(t.Format(time.ANSIC)))
	}
	pack.session.ChannelMessageSend(pack.message.ChannelID, message.String())
}

func (pc *ProfileCommand) getPermissionLevel(pack *CommPackage) string {
	// special checks for certain roles that aren't in the database
	if pack.message.Author.ID == pc.MasterId {
		return db.SprintPermission(db.PermMaster)
	} else if permissions.IsGuildOwner(pack.guild, pack.message.Author.ID) {
		return db.SprintPermission(db.PermGuildOwner)
	}

	perms := db.RoleQueryPermission(pack.member.Roles)
	highestPerm := db.PermAll
	// Find the highest permission level this user has
	for _, userPerm := range perms {
		if userPerm > highestPerm {
			highestPerm = userPerm
		}
	}
	return db.GetPermissionString(highestPerm)
}

func convertRankToString(rank int, serverMax sql.NullInt64) (rankString string) {
	if !serverMax.Valid {
		// no server max? just give back the rank itself
		return strconv.Itoa(rank)
	}
	// naming strategy: every 1% till 10%, then every 2% until 30%, then every 3% until 60% then every 4% until 100%, then every 100% forever
	rankPrefixes := []string{"Newcomer", "Apprentice", "Rookie", "Regular", "Veteran"}
	rankSeparator := " --> "
	percent := float64(rank) / float64(serverMax.Int64) * 100.0
	var rankPrefixIndex int
	var rankSuffix int
	if percent < 10 {
		rankSuffix = int(percent / 2.0)
		rankPrefixIndex = 0
	} else if percent < 30 {
		rankSuffix = int((percent - 10) / 4.0)
		rankPrefixIndex = 1
	} else if percent < 60 {
		rankSuffix = int((percent - 30) / 6.0)
		rankPrefixIndex = 2
	} else if percent < 100 {
		rankSuffix = int((percent - 60) / 8.0)
		rankPrefixIndex = 3
	} else {
		rankSuffix = int((percent - 100) / 100)
		rankPrefixIndex = 4
	}
	if rankSuffix != 0 {
		rankPrefixes[rankPrefixIndex] = util.MakeStringBold(rankPrefixes[rankPrefixIndex] + " " + strconv.Itoa(rankSuffix))
	} else {
		rankPrefixes[rankPrefixIndex] = util.MakeStringBold(rankPrefixes[rankPrefixIndex])
	}
	return convertToEmphasizedRankString(rankPrefixes, rankPrefixIndex, rankSeparator)
}

/*
Converts an array of strings to an emphasized string, currently used only for ranks. Looks like:
~element1~,**element2**, element3, element4
*/
func convertToEmphasizedRankString(ranks []string, indexToApply int, sep string) string {
	var s string
	for index, rankString := range ranks {
		if index < indexToApply {
			s += util.MakeStringStrikethrough(rankString)
		} else {
			s += rankString
		}
		if index < len(ranks)-1 {
			s += sep
		}
	}
	return s
}

func (pc *ProfileCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (pc *ProfileCommand) GetCommandKeys() []string {
	return []string{"PROFILE"}
}

func (pc *ProfileCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s profile` - Displays your server profile", commPrefix)
}
