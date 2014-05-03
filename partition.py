import sys, os, shutil, math

# sys.argv[1]		folder to partition
# sys.argv[2]		size of partition

def get_directories(path):
	return [name for name in os.listdir(path) if os.path.isdir('%s/%s' % (path, name))]	

def get_next_directory(dir_list):
	if len(dir_list) is 0:
		return 0
	else:
		return max( [int(name) for name in dir_list] ) + 1

partition_size = int(sys.argv[2])
directories = get_directories(sys.argv[1])
files = list( set(os.listdir(sys.argv[1])) - set(directories) )

next_directory_id = get_next_directory(directories)

count = 0
target_partition_count = math.ceil(len(files) / partition_size)
print 'Partitioning %d files into %d slices...' % (len(files), target_partition_count)
while len(files) > 0:
	folder_name = str(next_directory_id).zfill(4) 
	os.mkdir( '%s/%s' % (sys.argv[1], folder_name) )
	
	for filename in files[:partition_size]:
		shutil.move('%s/%s' % (sys.argv[1], filename), 
					'%s/%s/%s' % (sys.argv[1], folder_name, filename))

	files = files[partition_size:]
	next_directory_id += 1
	count += 1
	
	print '[Partition: %d / %d]\r' % (count, target_partition_count)
