import sys, os, re
import lib.GameRecord_pb2 as proto

# Read all files in a folder that match (*.gamelog).
# Read in and merge into a single log.

files = [fn for fn in os.listdir(sys.argv[1]) if fn.endswith('.gamelog')]
print files

gamelog = proto.GameLog()

print 'Merging %d files...' % len(files)
for i, fn in enumerate(files):
	with open( '%s/%s' % (sys.argv[1], fn) ) as fp:
		content = fp.read()
	
	partial = proto.GameLog()
	partial.ParseFromString(content)
	
	gamelog.games.extend( partial.games )
	print '[Merge: %d / %d]\r' % (i + 1, len(files))

with open('%s/master.gamelog' % sys.argv[1], 'w') as fp:
	fp.write(gamelog.SerializeToString())

print 'Complete!'
