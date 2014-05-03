import lib.GameRecord_pb2 as proto
import sys, bs4, time, re, os, time

## This is a fast and extraordinarily hacky parser that can be used to 
## quickly extract player ID and champion name from a list of scraped
## pages.

CHAMPION_PATTERN = re.compile("\"/champions/([a-z]+)\"")
SUMMONER_PATTERN = re.compile("\"/summoner/na/([0-9]+)\"")

# Gets a path and lists all subdirectories in the path.
def get_directories(path):
	return [name for name in os.listdir(path) if os.path.isdir('%s/%s' % (path, name))]	

# Extract all of the game information from the file. This often includes
# multiple games.
def parse(html):
	games = []
	
	# For each instance of phrase "winning team"
	while html.find("Winning Team") is not -1:
		record = proto.GameRecord()
		
		winning_team_loc = html.find("Winning Team")
		losing_team_loc = html.find("Losing Team")

		# Find the next five copies of champion URL's.
		# TODO(luke): note that this won't correctly reject games that aren't 5v5.
		#    Consider looking for "5v5" string near the winning team string?
		winning_champions = re.findall(CHAMPION_PATTERN, html[winning_team_loc:])[:5]
		losing_champions = re.findall(CHAMPION_PATTERN, html[losing_team_loc:])[:5]
		winning_summoners = re.findall(SUMMONER_PATTERN, html[winning_team_loc:])[:5]
		losing_summoners = re.findall(SUMMONER_PATTERN, html[losing_team_loc:])[:5]

		# Form the winners into a single team.
		winners = proto.Team()
		winners.victory = True
		for champion, summoner in zip(winning_champions, winning_summoners):
			player_stats = proto.PlayerStats()
			player_stats.player.name = champion
			player_stats.player.lolking_id = int(summoner)
			
			winners.players.extend([player_stats,])

		# Form the losers into a single team.
		losers = proto.Team()
		losers.victory = False
		for champion, summoner in zip(losing_champions, losing_summoners):
			player_stats = proto.PlayerStats()
			player_stats.player.name = champion
			player_stats.player.lolking_id = int(summoner)
			
			losers.players.extend([player_stats,])
		
		record.teams.extend([winners, losers])
		games.append(record)
		
		# Pop the used characters off the front of the string and keep
		# looking.
		html = html[max(winning_team_loc, losing_team_loc) + 12:]

	return games

directories = get_directories(sys.argv[1])

for i, directory in enumerate(directories):
	log = proto.GameLog()

	try:
		full_path = '%s/%s' % (sys.argv[1], directory)
		files = os.listdir(full_path)
		num_files = len(files)

		for fn in files:
			with open('%s/%s' % (full_path, fn)) as fp:
#				start = time.time()
				games = parse( fp.read() )
				log.games.extend(games)

#				print 'Progress: [%d / %d] (%.2fs)\r' % (i + 1, num_files, time.time() - start)
	finally:
		with open('%s/%s.gamelog' % (sys.argv[1], directory), 'wb') as fp:
			fp.write(log.SerializeToString())
		print "[Process: %d / %d]\r" % (i + 1, len(directories))
