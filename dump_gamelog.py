import lib.GameRecord_pb2 as proto
import sys

gamelog = proto.GameLog()

with open(sys.argv[1]) as fp:
	gamelog.ParseFromString( fp.read() )

for game in gamelog.games:
	out = []

#	if len(game.teams) < 2:
#		raise Exception("one team only!")
#	out.append("V" if game.teams[0].victory else "L")
#	out.append("V" if game.teams[1].victory else "L")
	
	for team in game.teams:
		out.append( ','.join([player.champion for player in team.players]))
		
	print '\t'.join(out)
