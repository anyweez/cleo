import lib.GameRecord_pb2 as proto
import sys, bs4, time, re, os

import time

## Accepts a directory and parses game information from every file in the specified directory.

## TODO: where to call to retrieve main player stats?

def parse(html):
	start = time.time()
	soup = bs4.BeautifulSoup(html, parse_only=bs4.SoupStrainer('div'))

	print "  [parse: %.2f]" % (time.time() - start)
	
	games = []

	for game in soup.find_all( 'div', { 'class': 'match_win'} ):
		record = parse_game(game)	

		if game.teams is not None and len(game.teams) > 1:
			games.append(record)
		
	for game in soup.find_all( 'div', { 'class': 'match_loss'} ):
		record = parse_game(game)

		if game.teams is not None and len(game.teams) > 1:
			games.append(record)
		
	return games

# Given a game element, extract all of the game-level attributes and return a GameRecord object.
def parse_game(element):
	record = proto.GameRecord()

#	record.lolking_id = int( element['data-game-id'] )
#	match = re.match('([0-9]+)/([0-9]+)/([0-9]+) ([A-Z0-9:]+)', element.find('span')['data-hoverswitch'])
#	month, day, year, rest = str(match.group(1)), str(match.group(2)), str(match.group(3)), match.group(4)
	
#	record.timestamp = int( time.mktime( time.strptime('%s/%s/%s %s' % (month.zfill(2), day.zfill(2), year.zfill(2), rest), '%m/%d/%y %I:%M%p') ) )
	
	# TODO: This is an unstable parse. Currently depends on the fact that no <strong>
	#   elements come before it.
#	try:
#		match = re.match('([0-9]+)[\+]* min', element.find('strong').getText())
#		record.duration =  int( match.group(1) )
		# 55+ has a different format. If there aren't any groups returned, 
		# assume 55+ minutes.
#	except AttributeError:
#		record.duration = 55
	
	for team in element.find_all('td'):
		if len( team.findChildren() ) > 0:
			extracted_team = parse_team(team)
			if len(extracted_team.players) > 0:
				record.teams.extend( [extracted_team,] )

	return record

# Given a team element, extract all of the champions involved in the game.
def parse_team(element):
	team = proto.Team()
	
#	victory_set = False
	
	players = []
	for plyr in element.find_all('tr'):
		player_stats = proto.PlayerStats()
		player = proto.Player()
		
		if re.match('Losing Team', plyr.getText()):
			team.victory = False
#			victory_set = True
		elif re.match('Winning Team', plyr.getText()):
			team.victory = True
#			victory_set = True
		else:
			player_stats.player.name = plyr.getText().strip()
			player_stats.champion = plyr.find('a')['href'].split('/')[-1]
			
#			try:
#				id_element = plyr.find('a', href=re.compile('/summoner/[a-z][a-z]/([0-9]+)'))
#				player_stats.player.lolking_id = int( id_element['href'].split('/')[-1] )
			# This is the player whose page the game was discovered on. Use data
			# read from elsewhere.
#			except TypeError:
				# Note: this probably isn't the most future-proof way to do this...
#				player_stats.player.lolking_id = int( sys.argv[1].split('/')[-1].split('.')[0] )
		
			players.append(player_stats)
	
	team.players.extend(players)
	
#	if not victory_set:
#		raise Exception("Team doesn't have known victory outcome.")
	
	return team

#def parse_player_stats(player, element):
#	kda_element = element.find_all('div')[3]
#	kda_each = kda_element.find_all('strong')
	
#	print 'kills: %d' % int( kda_each[0].getText() )
#	player.kills = int( kda_each[0].getText() )
#	player.deaths = int( kda_each[1].getText() )
#	player.assists = int( kda_each[2].getText() )
#	player.gold
#	player.minions
#
#	return player
	
# Merges a list of games into a single representation.
# Input: a list of games
# Output: a single game that has merged PlayerStats for all available players.
#   GameStats should not change.
def merge(games):
	pass

log = proto.GameLog()

files = os.listdir(sys.argv[1])
num_files = len(files)
print 'Directory contains %d files' % num_files

try:
	for i, fn in enumerate(files):
		with open('%s/%s' % (sys.argv[1], fn)) as fp:
			start = time.time()
			games = parse( fp.read() )
			log.games.extend(games)

			print 'Progress: [%d / %d] (%.2fs)\r' % (i + 1, num_files, time.time() - start)
finally:
	print "Writing to %s.gamelog..." % sys.argv[1] 
	with open('%s.gamelog' % sys.argv[1], 'wb') as fp:
		fp.write(log.SerializeToString())
	print "Complete."
