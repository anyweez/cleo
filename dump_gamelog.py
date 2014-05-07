import lib.gamelog_pb2 as proto
import sys

gamelog = proto.GameLog()

with open(sys.argv[1]) as fp:
	gamelog.ParseFromString( fp.read() )

print 'Games count: %d' % len(gamelog.games)
for game in gamelog.games:
	out = []

#	if len(game.teams) < 2:
#		raise Exception("one team only!")
#	out.append("V" if game.teams[0].victory else "L")
#	out.append("V" if game.teams[1].victory else "L")

	print "Game %d, teams: %d" % (game.game_id, len(game.teams))
	for team in game.teams:
		print '  ' + ','.join([str(player.champion_id) for player in team.players])
		
