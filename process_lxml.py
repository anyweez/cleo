import lib.GameRecord_pb2 as proto
from lxml import etree
import os, sys

## Accepts a directory and parses game information from every file in the specified directory.

## TODO: where to call to retrieve main player stats?

def parse(html):
	root = etree.fromstring(html)
        games = []

#        for game in soup.find_all( 'div', { 'class': 'match_win'} ):
#                record = parse_game(game)

#                if game.teams is not None and len(game.teams) > 1:
#                        games.append(record)

#		for game in soup.find_all( 'div', { 'class': 'match_loss'} ):
#		        record = parse_game(game)

#			if game.teams is not None and len(game.teams) > 1:
#				games.append(record)

        return games

## Main codepath.
log = proto.GameLog()

files = os.listdir(sys.argv[1])
num_files = len(files)
print 'Directory contains %d files' % num_files

try:
	for i, fn in enumerate(files):
		print "reading %s" % fn
		
		with open('%s/%s' % (sys.argv[1], fn)) as fp:
			games = parse( fp.read() )
			log.games.extend(games)

		print 'Progress: [%d / %d]\r' % (i + 1, num_files)
finally:
	print "Writing to %s.gamelog..." % sys.argv[1]
	with open('%s.gamelog' % sys.argv[1], 'wb') as fp:
		fp.write(log.SerializeToString())
	
	print "Complete."

