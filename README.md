This application is the analogue of unix pipeline, sort of:
```
grep 127.0.0.1 | awk '{print $2}' | sort | uniq -c | sort -nr
```

When STDOUT of one application is transmitted as STDIN in another application.

But in our case, these roles are performed by channels that we transfer from one function to another.

Functions can calculate hash-sum of given array of integers and do other things.

The calculation of the hash-sum is implemented by the following chain:
* SingleHash calculates value crc32(data)+"~"+crc32(md5(data)) (concatenation of two strings by ~), where data - what came from STDIN.
* MultiHash calculates value crc32(th+data)) (concatenation of digit, parsed to string, and string), where th=0..5 (i.e. 6 hashes by every input value ), then takes concatenations of results in the order of calculation (0..5), where data - what came from STDIN (STDOUT from SingleHash).
* CombineResults takes all results, sorts them, concatenates the sorted result through '_' in one string.
* crc32 calculates by DataSignerCrc32 function.
* md5 calculates by DataSignerMd5.

What's the catch:
* DataSignerMd5 can only be called at the same time, calculation takes 10 ms. If several start at the same time - there will be overheat by 1 second.
* DataSignerCrc32, calculation takes 1 second.
* For all calculations we have 3 seconds.
* If program will work linear, for 7 elements it will take almost 57 seconds.

Results, which are displayed if you send 2 values:

```
0 SingleHash data 0
0 SingleHash md5(data) cfcd208495d565ef66e7dff9f98764da
0 SingleHash crc32(md5(data)) 502633748
0 SingleHash crc32(data) 4108050209
0 SingleHash result 4108050209~502633748
4108050209~502633748 MultiHash: crc32(th+step1)) 0 2956866606
4108050209~502633748 MultiHash: crc32(th+step1)) 1 803518384
4108050209~502633748 MultiHash: crc32(th+step1)) 2 1425683795
4108050209~502633748 MultiHash: crc32(th+step1)) 3 3407918797
4108050209~502633748 MultiHash: crc32(th+step1)) 4 2730963093
4108050209~502633748 MultiHash: crc32(th+step1)) 5 1025356555
4108050209~502633748 MultiHash result: 29568666068035183841425683795340791879727309630931025356555

1 SingleHash data 1
1 SingleHash md5(data) c4ca4238a0b923820dcc509a6f75849b
1 SingleHash crc32(md5(data)) 709660146
1 SingleHash crc32(data) 2212294583
1 SingleHash result 2212294583~709660146
2212294583~709660146 MultiHash: crc32(th+step1)) 0 495804419
2212294583~709660146 MultiHash: crc32(th+step1)) 1 2186797981
2212294583~709660146 MultiHash: crc32(th+step1)) 2 4182335870
2212294583~709660146 MultiHash: crc32(th+step1)) 3 1720967904
2212294583~709660146 MultiHash: crc32(th+step1)) 4 259286200
2212294583~709660146 MultiHash: crc32(th+step1)) 5 2427381542
2212294583~709660146 MultiHash result: 4958044192186797981418233587017209679042592862002427381542

CombineResults 29568666068035183841425683795340791879727309630931025356555_4958044192186797981418233587017209679042592862002427381542
```

Run as `go test -v`