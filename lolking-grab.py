import sys, urllib2, bs4, re, datetime, os, random
from collections import deque

#champions = {}
#newly_discovered = {}
champion_queue = deque()
champ_map = {}

# Write out new champion list
def write_champions(out_queue):
	unicode_count = 0
	iterations = 0
	print 'Writing %d champions to champions.out...' % len(out_queue)
	with open('champions.out', 'w') as fp:
		for name, url, last_retrieved in out_queue:
			try:
				fp.write( '%s\t%s\t%s\n' % (name, url, last_retrieved) )
				iterations += 1
			except UnicodeEncodeError:
				unicode_count += 1
	
	print '  iterations: %d' % iterations
	print '  (skipped %d unicode character names)' % unicode_count

### Stage 1: read in champion list. We'll retrieve the match page for
### each champion in the provided list.

# Read in the champion list
with open(sys.argv[1]) as fp:
	champ_list = fp.readlines()
	for champ in champ_list:
		try:
			name, url, last_retrieved = champ.split('\t')
			champion_queue.append( [name.strip(), url.strip(), last_retrieved.strip()] )
			champ_map[name.strip()] = True
		except ValueError:
			print 'Invalid line. Skipping...'

random.shuffle(champion_queue)

### Stage 2: Retrieve champion pages, save them, and extract new
### champion names.
try:
	while True:
		output_folder = '/mnt/vortex/corpora/lolking/%s' % datetime.datetime.now().strftime('%Y-%m-%d')
		try:
			os.makedirs( output_folder )
		except OSError:
			pass

		# Pop this champ
		name, url, last_retrieved = champion_queue.popleft()
		del champ_map[name]
	
		print 'Retrieving %s [%s] | Queue length: %d' % ( name, url, len(champion_queue) )
		response = urllib2.urlopen(url)

		with open( '%s/%s.html' % (output_folder, url.split('/')[-1]), 'wb') as fp:
			# Write it to a file.
			html = response.read()
			fp.write(html)
		
			# Extract new champion names.
			soup = bs4.BeautifulSoup(html)
			for element in soup.find_all( 'a', href=re.compile('/summoner/na/[0-9]+'), text=True, ):
				print '\t %s [http://www.lolking.net%s]' % (element.getText(), element['href'])
			
				# Only add the champion to the queue if they don't already exist there.
				if not champ_map.has_key( element.getText() ):
					champion_queue.append( [element.getText(), 'http://www.lolking.net%s' % element['href'], 0] )
					champ_map[element.getText()] = True
except:
	write_champions(champion_queue)
	raise
