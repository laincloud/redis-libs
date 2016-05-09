import random
import redis
import logging
import time

REDIS_HOST = '127.0.0.1'
REDIS_PORT = 7001
DEBUG = 1
TIMEOUT = 10

r = redis.StrictRedis(host=REDIS_HOST,port=REDIS_PORT,socket_timeout=TIMEOUT)


def write_data_with_redis_client(*args):
    try:
        global key
        key = random.randint(1,100000)
        value = random.randint(1,100000)
        res = str(r.set(key,value))
        print 'redisclient set '+ str(key) + '\'s value: '+ str(value) + " : "+ str(res)
    except Exception as e:
        logging.error(e)
        pass

if __name__ == "__main__":
    while True:
        write_data_with_redis_client()
        time.sleep(1)
