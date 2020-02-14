#!/usr/bin/env python

import random
import uuid
import requests
import time
from datetime import datetime
import sys

prefixes = ('9023',)#, '9024', '9025')
codes =('220', '22020', '22021', '22022', '2203', '2206', '2207', '2208', '2209', '22177', '22178', '2237', '22390', '22391', '22392', '22393', '22501', '22502', '22503', '22504', '22505', '22506', '22507', '22508', '22509', '22540', '22541', '22542', '22544', '22545', '22546', '22547', '22548', '22549', '22555', '22556', '22557', '22558', '22559', '22565', '22566', '22567', '22577', '2314', '2315', '2316', '2317', '23188', '23199', '23277', '23288', '235222', '235223', '235224', '235227', '235228', '2353', '2356', '2357', '2359', '25215', '25224', '25228', '25229', '25250', '25251', '25259', '25260', '25261', '25290', '25291', '25299', '255', '2557', '25777', '26377', '26378', '35568', '35569', '37740', '37741', '37742', '37743', '37744', '37745', '37746', '37747', '37748', '37749', '3816', '38160', '38161', '38162', '38163', '38164', '38165', '38166', '38168', '38169', '3897', '50931', '50936', '50937', '50938', '50946', '50947', '50948')

url = "http://localhost:2080/freeswitch_json"
headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}

start_time = datetime.today().timestamp()
i = 0

with open('Generator_json.txt') as f:
    text = f.read()
    loop_forever = True
    loop = 0
    if len(sys.argv) > 1:
        loop_forever = False
        loop = int(sys.argv[1])
    loop_index = 0
    while loop_forever or loop_index < loop:
        usage = random.randrange(18000)
        data = text % dict(billsec=usage, billmsec = usage*1000, uuid=str(uuid.uuid4()),
                           sip_to_user='%s%s%d' % (random.choice(prefixes), random.choice(codes), random.randrange(999999999)))
        r = requests.post(url, data=data, headers=headers)
        time_diff = datetime.today().timestamp() - start_time
        i += 1
        if i%100 == 0:
            print(int(i / time_diff),"req/s")
        loop_index += 1
